#include "runtime.h"
extern byte _MaximizePng[], _eMaximizePng;

void ·getMaximizePng(Slice s) {
    s.array = _MaximizePng;
    s.len = s.cap = &_eMaximizePng - _MaximizePng;
    FLUSH(&s);
}
