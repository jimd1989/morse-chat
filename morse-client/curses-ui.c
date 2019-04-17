#include "curses-ui.h"

void initScreen(Screen *s, unsigned int *audioOn) {
    initscr();
    noecho();
    raw();
    mousemask(BUTTON1_PRESSED | BUTTON1_RELEASED, NULL);
    mouseinterval(0);
    keypad(stdscr, TRUE);
    scrollok(stdscr, TRUE);
    s->audioOn = audioOn;
}

/* Screen.ch will be checked fully in Go. The purpose of this function is to
 * short circuit the on/off process */

void getInput(Screen *s) {
    s->ch = getch();
    if (s->ch == KEY_MOUSE) {
        if (getmouse(&s->event) == OK) {
            if (s->event.bstate & BUTTON1_PRESSED) {
                s->ch = 111; /* the 'o' (on) key */
                *s->audioOn = 1;
            } else {
                s->ch = 112; /* the 'p' (pff?) key */
                *s->audioOn = 0;
            }
        }
    }
}

void cursesPrintln(const char *s) {
    printw("%s\n", s);
    refresh();
}

/* Grabs typed in user values and returns an actual number. Used to set hz and
 * volume */

double getText() {
    echo();
    char buffer[STR_MAX] = { 0 };
    double d;
    int i, ch;
    i = ch = 0;
    while (i < STR_MAX - 1 && ch != '\n') {
        ch = getch();
        buffer[i++] = ch;
    }
    noecho();
    sscanf(buffer, "%lf", &d);
    if (i == STR_MAX - 1) {
        printw("\n");
        refresh();
    }
    return d;
}
