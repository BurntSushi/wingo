package main

var newx, newy int16 // prevent memory allocation in 'step' functions

func (f *frameNada) moveBegin(rx, ry, ex, ey int16) {
    f.moving.lastRootX, f.moving.lastRootY = rx, ry

    // call for side-effect; makes sure parent window has a valid geometry
    f.parent.window.geometry()
}

func (f *frameNada) moveStep(rx, ry, ex, ey int16) {
    newx = f.parent.window.geom.X() + rx - f.moving.lastRootX
    newy = f.parent.window.geom.Y() + ry - f.moving.lastRootY
    f.moving.lastRootX, f.moving.lastRootY = rx, ry

    f.configureFrame(DoX | DoY, newx, newy, 0, 0, 0, 0)
}

func (f *frameNada) moveEnd(rx, ry, ex, ey int16) {
    f.moving.lastRootX, f.moving.lastRootY = 0, 0
}

func (f *frameNada) resizeBegin(rx, ry, ex, ey int16) {
    logMessage.Printf("resize begin: %s", f.client())
}

func (f *frameNada) resizeStep(rx, ry, ex, ey int16) {
    logMessage.Printf("resize step: %s", f.client())
}

func (f *frameNada) resizeEnd(rx, ry, ex, ey int16) {
    logMessage.Printf("resize end: %s", f.client())
}

