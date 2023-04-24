#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/stat.h>
#include <libetpan/libetpan.h>
#include <signal.h>
#include <unistd.h>
#include <fcntl.h>
#include <time.h>

/*
 * set the mail, password, mail server and folders
 * dependency: net-libs/libetpan
 */

char user_mail[] = "user1@mail.com\0";
char user_pass[] = "pass1\0";
char mail_server[] = "imap.mail.com\0";
int mail_port = 993;

char *mbox_list[] = {	"INBOX\0",
			"INBOX/folder1\0",
			"INBOX/folder2\0",
			"Sent\0",
			"\0"};
/*
 * gmail - example folders
 * { "INBOX", "[Gmail]/All Mail", "[Gmail]/Sent Mail", 
 * "[Gmail]/Drafts", "[Gmail]/Spam", "[Gmail]/Bin", "\0"};
 */


char new_mail[] = "0";

#define MSG_SIZE 1024
#define BUF_SIZE 8

struct mail_st_s {
	uint32_t all;
	uint32_t recent;
	uint32_t unseen;
};

static void check_error(int r, char *msg);
void sig_handler(int signo);
void get_status(struct mailimap *imap, struct mail_st_s *mail_st, char *mb);
void prepare_st(struct mailimap *imap, struct mail_st_s *mail_st,
		char **mbox_list, char *msg);

int main(int argc, char **argv) {
	struct mailimap *imap;
	int r;
	char msg[MSG_SIZE] = {};

	struct mail_st_s *mail_st = malloc(sizeof(*mail_st));
	memset(mail_st, 0, sizeof(struct mail_st_s));

	if (argc > 2) 
		printf("Too many arguments.\nUsage: mailcount <-d>\n");

	if (signal(SIGINT, sig_handler) == SIG_ERR)
		printf("\ncan't catch signal\n");

	imap = mailimap_new(0, NULL);
	r = mailimap_ssl_connect(imap, mail_server, mail_port);
	check_error(r, "could not connect to server");

	r = mailimap_login(imap, user_mail, user_pass);
	check_error(r, "could not login");

	if (argc==2 && !strcmp(argv[1], "-d")) {
		printf("running in daemon mode...\n");

		daemon(0, 0);

		int file_fd;
		char file_name[] = "/tmp/new_mail";
		struct stat st;

		// if no file - create it
		if (stat(file_name, &st) != 0) {
			FILE *fp = NULL;
			fp = fopen(file_name ,"w");
			chmod(file_name, 0644);
			fclose(fp);
		}

		// start sending messages
		while (1) {
			*new_mail = '0';
			for (int i = 0; *mbox_list[i] != '\0'; i++) {
				get_status(imap, mail_st, mbox_list[i]);
			}

			file_fd = open(file_name, O_WRONLY);
			write(file_fd, new_mail, 1);
			close(file_fd);
		}
	} else {
		prepare_st(imap, mail_st, mbox_list, msg);
		printf("new mail: %s\n\n", new_mail);
		printf("%s", msg);
	}

	mailimap_logout(imap);
	mailimap_free(imap);

	exit(0);
}

static void check_error(int r, char *msg) {
	if (r == MAILIMAP_NO_ERROR)
		return;
	if (r == MAILIMAP_NO_ERROR_AUTHENTICATED)
		return;
	if (r == MAILIMAP_NO_ERROR_NON_AUTHENTICATED)
		return;

	fprintf(stderr, "%s\n", msg);
	exit(1);
}

void sig_handler(int signo) {
	if (signo == SIGINT) {
		printf("\nreceived SIGINT; terminating.\n");
		exit(0);
	}
}

void get_status(struct mailimap *imap, struct mail_st_s *mail_st, char *mb) {
	struct mailimap_status_att_list *att_list;
	struct mailimap_mailbox_data_status *data_status;
	struct mailimap_status_info * status_info;

	att_list = mailimap_status_att_list_new_empty();
	mailimap_status_att_list_add(att_list, MAILIMAP_STATUS_ATT_MESSAGES);
	mailimap_status_att_list_add(att_list, MAILIMAP_STATUS_ATT_RECENT);
	mailimap_status_att_list_add(att_list, MAILIMAP_STATUS_ATT_UNSEEN);

	mailimap_status(imap, mb, att_list, &data_status);

	clistiter * cur;

	for (cur = clist_begin(data_status->st_info_list);
			cur != NULL;
			cur = clist_next(cur)) {

		status_info = clist_content(cur);
		switch (status_info->st_att) {
			case MAILIMAP_STATUS_ATT_MESSAGES:
				mail_st->all = status_info->st_value;
				break;
			case MAILIMAP_STATUS_ATT_RECENT:
				mail_st->recent = status_info->st_value;
				break;
			case MAILIMAP_STATUS_ATT_UNSEEN:
				mail_st->unseen = status_info->st_value;
				if (mail_st->unseen)
					*new_mail = '1';
				break;
		}
	}

	mailimap_status_att_list_free(att_list);
	mailimap_mailbox_data_status_free(data_status);
	//mailimap_status_info_free(status_info);
	free(cur);
}

void prepare_st(struct mailimap *imap, struct mail_st_s *mail_st,
		char **mbox_list, char *msg) {

	*msg = '\0';

	for (int i = 0; *mbox_list[i] != '\0'; i++) {
		char buf[BUF_SIZE];

		get_status(imap, mail_st, mbox_list[i]);

		strncat(msg, mbox_list[i], MSG_SIZE);
		strncat(msg, " - messages: ", MSG_SIZE);
		snprintf(buf, BUF_SIZE, "%d", mail_st->all);
		strncat(msg, buf, MSG_SIZE);
		strncat(msg, "\n", MSG_SIZE);

		strncat(msg, mbox_list[i], MSG_SIZE);
		strncat(msg, " - recent: ", MSG_SIZE);
		snprintf(buf, BUF_SIZE, "%d", mail_st->recent);
		strncat(msg, buf, MSG_SIZE);
		strncat(msg, "\n", MSG_SIZE);

		strncat(msg, mbox_list[i], MSG_SIZE);
		strncat(msg, " - unseen: ", MSG_SIZE);
		snprintf(buf, BUF_SIZE, "%d", mail_st->unseen);
		strncat(msg, buf, MSG_SIZE);
		strncat(msg, "\n", MSG_SIZE);

		strncat(msg, "\n", MSG_SIZE);
	}
}
