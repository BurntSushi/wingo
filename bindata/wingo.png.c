#include "runtime.h"
extern byte _WingoPng[], _eWingoPng;

void ·getWingoPng(Slice s) {
    s.array = _WingoPng;
    s.len = s.cap = &_eWingoPng - _WingoPng;
    FLUSH(&s);
}
