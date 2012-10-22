#include "runtime.h"
extern byte _DejavusansTtf[], _eDejavusansTtf;

void ·getDejavusansTtf(Slice s) {
    s.array = _DejavusansTtf;
    s.len = s.cap = &_eDejavusansTtf - _DejavusansTtf;
    FLUSH(&s);
}
