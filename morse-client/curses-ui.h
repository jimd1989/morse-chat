#include <curses.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

/* The client uses a primitive curses interface, mainly to retrieve mouse/key
 * events and send Msgs from them. */

#define STR_MAX 16

/* The Screen type contains a pointer for mouse events, as well as a pointer
 * directly to the client's AudioInstance.on value, so that sound may be
 * rendered ASAP. All other (non time sensitive) events are routed through
 * Msgs to the server */

typedef struct Screen {
    int ch;
    unsigned int *audioOn;
    MEVENT event;
} Screen;

void initScreen(Screen *, unsigned int *);
void getInput(Screen *);
void cursesPrintln(const char *);
double getText(void);


