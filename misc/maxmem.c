#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <unistd.h>

#define MAXCHAR 32


long get_val(char *line_in);
int starts_with(const char *str, const char *start_str);


int main(void)
{
	FILE *fp;
	char *filename = "/proc/meminfo";

	char line[MAXCHAR];
	long mem_used, mem_all_buff, mem_used_min=999999999, mem_used_max=0;
	int count;

	char *mem_items[] = { 	"MemTotal:",	"MemAvailable:",
				"MemFree:",	"Buffers:",
				"Cached:",	"SReclaimable:" };

	int mem_items_len = sizeof(mem_items) / sizeof(mem_items[0]);

	long mem_vals[mem_items_len]; 
	
	for (;;) {

	count = 0;
	fp = fopen(filename, "r");

	if (fp == NULL) {
		printf("Could not open file %s",filename);
		return 1;
	}


	while (fgets(line, MAXCHAR, fp) != NULL) {
	
		for (int i=0; i < mem_items_len; i++)
			if (starts_with(line, mem_items[i])) {
				mem_vals[i] = get_val(line);
				count++;
				break;
			}

		if (count == mem_items_len)
			break;
	}

	fclose(fp);

	mem_all_buff = mem_vals[3] + mem_vals[4] + mem_vals[5];
	mem_used = mem_vals[0] - mem_vals[2] - mem_all_buff;

	if (mem_used < mem_used_min)
		mem_used_min = mem_used;
		
	if (mem_used > mem_used_max)
		mem_used_max = mem_used;

	printf("Total:             %6.d\n", mem_vals[0] / 1024);
	printf("Available:         %6.d\n", mem_vals[1] / 1024);
	printf("Free:              %6.d\n", mem_vals[2] / 1024);
	printf("Cached/Buffers:    %6.d\n", mem_all_buff / 1024);
	printf("Used:              %6.d\n", mem_used / 1024);
	printf("                       \n");
	printf("Used min:          %6.d\n", mem_used_min / 1024);
	printf("Used max:          %6.d\n", mem_used_max / 1024);
	printf("\033[8A");
	
	usleep(100000);

	}

	return 0;
}


long get_val(char *line_in)
{
	char val[MAXCHAR];

	while (! isdigit(*line_in) && ! isspace(*line_in++ - 1))

		strcpy(val, line_in);


	for (char *p = val; *p != '\n'; p++)

		if (isspace(*p)) {
			*p = 0;
			break;
		}
	
	return atoi(val);
}

int starts_with(const char *str, const char *start_str)
{
	return strncmp(start_str, str, strlen(start_str)) == 0;
}

