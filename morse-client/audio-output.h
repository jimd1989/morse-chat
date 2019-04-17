#include <limits.h>
#include <math.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#include <ao/ao.h>

/* The main synthesis and audio output loop. Every User has a corresponding
 * AudioInstance struct, which contains his/her note status and frequency info.
 * All instances, along with other playback info, are stored in an Out struct,
 * which is exposed to the Go code. Audio playback runs constantly in its own
 * goroutine. Updates to playback are made through direct atomic changes to
 * the values of the Out struct. */

#define RATE 48000
#define RESOLUTION 96
#define BUFFSIZE (RATE / RESOLUTION)
#define O_BUFFSIZE (BUFFSIZE * 2)
#define WAVELEN 4096
#define TWOPI (2.0 * M_PI)
#define SINE_INCREMENT (TWOPI / (double)WAVELEN)
#define EVENT_INCREMENT ((double)WAVELEN / (double)RATE)

/* The AudioInstance type contains a User's relevant playback information. This
 * struct is referenced and updated while filling the audio buffer. The
 * instance's waveform is synthesized regardless of whether or not the User has
 * audio on (1) or off (0). This arrangement allows the User to update the
 * AudioInstance.on field atomically in the middle of a buffer filling
 * operation, which ensures high click resolution regardless of buffer size.
 * New pitches are written to AudioInstance.newPitch, which is checked between
 * buffer fills and updated accordingly, avoiding the need for mutexes. */

typedef struct AudioInstance {
    unsigned int on;
    double newPitch;
    double phase;
    double pitch;
} AudioInstance;

/* The Out type is a Go-facing struct that contains all playback information.
 * It is meant to be stack allocated. Check the values of BUFFSIZE and
 * O_BUFFSIZE if this presents a problem. */

typedef struct Out {
    uint32_t phase;
    int active;
    unsigned int usersMax;
    double masterAmplitude;
    double mixAmplitude;
    AudioInstance *instances;
    double wave[WAVELEN];
    char buffer[O_BUFFSIZE];
    double mixer[BUFFSIZE];
    ao_device *device;
    ao_sample_format format;
    int default_driver;
} Out;

void initWave(double *);
int initOut(Out *, const unsigned int);
void destroyOut(Out *);
AudioInstance * getInstance(Out *, const unsigned int);
void playback(Out *);
void changeOn(Out *, int, int);
void changePitch(Out *, int, double);
