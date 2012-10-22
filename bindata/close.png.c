#include "runtime.h"
extern byte _ClosePng[], _eClosePng;

void ·getClosePng(Slice s) {
    s.array = _ClosePng;
    s.len = s.cap = &_eClosePng - _ClosePng;
    FLUSH(&s);
}
