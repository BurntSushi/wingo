#include "runtime.h"
extern byte _WingoWav[], _eWingoWav;

void ·getWingoWav(Slice s) {
    s.array = _WingoWav;
    s.len = s.cap = &_eWingoWav - _WingoWav;
    FLUSH(&s);
}
