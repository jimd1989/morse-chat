#include "audio-output.h"

void initWave(double *w) {
    int i;
    double p;
    for (i = 0, p = 0.0f ; i < WAVELEN ; i++, p += SINE_INCREMENT) {
        w[i] = sinf(p);
    }
}

int initOut(Out *o, const unsigned int usersMax) {
    o->phase = 0;
    o->active = 1;
    o->usersMax = usersMax;
    o->masterAmplitude = 1.0; 
    o->mixAmplitude = 0.95 / (double)o->usersMax;
    o->instances = calloc(o->usersMax, sizeof(*o->instances));
    if (o->instances == NULL) {
        fprintf(stderr, "Error allocating memory for audio instances.\n");
        return -1;
    }
    initWave(o->wave);
    memset(o->buffer, 0, O_BUFFSIZE * sizeof(char));
    memset(o->mixer, 0, BUFFSIZE * sizeof(double));
    ao_initialize();
    o->default_driver = ao_default_driver_id();
    memset(&o->format, 0, sizeof(o->format));
    o->format.bits = 16;
    o->format.channels = 1;
    o->format.rate = 48000;
    o->format.byte_format = AO_FMT_LITTLE;
    o->device = ao_open_live(o->default_driver, &o->format, NULL);
    if (o->device == NULL) {
        fprintf(stderr, "Error opening device.\n");
        return -1;
    }
    return 0;
}

void destroyOut(Out *o) {
    if (o->instances == NULL) {
        return;
    }
    free(o->instances);
}

/* Returns the address of a specific instance, so that it is directly
 * accessible by Go. */

AudioInstance *getInstance(Out *o, const unsigned int i) {
    return &o->instances[i];
}

/* Main playback loop. Sines are synthesized from simple truncating wavetable
 * lookup, since audio fidelity is not a concern with something like morse.
 * Runs in a single thread for the time being. The algorithm is trivial to
 * parallelize, but the lack of pthread barriers on macOS makes it more trouble
 * than it's worth. The Out.phase field is capable of overflowing (after a 
 * very long time.) This is acceptable. */

void playback(Out *o) {
    unsigned int i, j;
    double d;
    int16_t b;
    AudioInstance *ai = NULL;
    while (o->active == 1) {
        o->phase++;
        memset(o->mixer, 0, BUFFSIZE * sizeof(double));
        for (i = 0 ; i < o->usersMax ; i++) {
            ai = &o->instances[i];
            if (ai->newPitch != 0.0) {
                ai->pitch = ai->newPitch * EVENT_INCREMENT;
                ai->phase = fmod(ai->pitch * (double)o->phase *
                                 (double)BUFFSIZE, TWOPI);
                ai->newPitch = 0.0;
            }
            for (j = 0 ; j < BUFFSIZE ; j++) {
                ai->phase += ai->pitch;
                d = o->wave[(unsigned int)ai->phase % WAVELEN];
                o->mixer[j] += d * o->mixAmplitude * ai->on;
            }
        }
        for (i = 0, j = 0 ; i < BUFFSIZE ; i++, j += 2) {
            b = (int16_t)(o->mixer[i] * o->masterAmplitude * SHRT_MAX);
            o->buffer[j] = (char)(b & 255);
            o->buffer[j+1] = (char)(b >> 8);
        }
        ao_play(o->device, o->buffer, O_BUFFSIZE);
    }
    ao_close(o->device);
    ao_shutdown();
}
