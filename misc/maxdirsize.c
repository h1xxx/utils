#include <stdio.h>
#include <errno.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <ftw.h>

static unsigned int total = 0;

int sum(const char *fpath, const struct stat *sb, int typeflag) {
	total += sb->st_size;

	return 0;

}

int main(int argc, char **argv) {
	if (!argv[1] || access(argv[1], R_OK)) {
		return 1;
	}

	if (ftw(argv[1], &sum, 1)) {
		perror("ftw");
		return 2;
	}

	printf("%s: %u\n", argv[1], total / 1024 / 1024);

	return 0;
}


/*
Here is a recursive version:

#include <unistd.h>
#include <sys/types.h>
#include <dirent.h>
#include <stdio.h>
#include <string.h>

void listdir(const char *name, int indent)
{
    DIR *dir;
    struct dirent *entry;

    if (!(dir = opendir(name)))
        return;

    while ((entry = readdir(dir)) != NULL) {
        if (entry->d_type == DT_DIR) {
            char path[1024];
            if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..")
                continue;
            snprintf(path, sizeof(path), "%s/%s", name, entry->d_name);
            printf("%*s[%s]\n", indent, "", entry->d_name);
            listdir(path, indent + 2);
        } else {
            printf("%*s- %s\n", indent, "", entry->d_name);
        }
    }
    closedir(dir);
}

int main(void) {
    listdir(".", 0);
    return 0;
}

*/
