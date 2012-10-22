#include "runtime.h"
extern byte _MinimizePng[], _eMinimizePng;

void ·getMinimizePng(Slice s) {
    s.array = _MinimizePng;
    s.len = s.cap = &_eMinimizePng - _MinimizePng;
    FLUSH(&s);
}
