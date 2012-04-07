package main

// renderTextSolid does the plumbing for drawing text on a solid background.
// It returns the width and height of the TEXT, the image itself may be bigger.
func renderTextSolid(bgColor int, font *truetype.Font, fontSize float64,
                     fontColor int, text string) (*wImg, int, int, error) {
    // The extents appear to cut off some of the text.
    // It may be a bad idea to hard code this value, but my knowledge
    // of drawing fonts is pretty shaky. It seems to work.
    breathe := 5

    // ew and eh are the *max* text extents (since it assumes every character
    // is 1em in width). I don't know how to get accurate text extents without
    // actually drawing the text, so this will have to due for now. We'll end
    // up creating bigger images than we need, but we can resize the window
    // itself after we get the *real* extents when we draw the text.
    ew, eh, err := xgraphics.TextMaxExtents(font, fontSize, text)
    if err != nil {
        logWarning.Printf("Could not get text extents for text '%s' " +
                          "because: %v", text, err)
        logWarning.Printf("Resorting to default with of 200.")
        ew = 200
    }

    // Draw the background for the text plus some breathing room
    textImg := renderSolid(bgColor, ew + breathe, eh + breathe)

    // rew and reh are the real text extents (since we started at 0, 0)
    rew, reh, err := xgraphics.DrawText(textImg, 0, 0, colorFromInt(fontColor),
                                        fontSize, font, text)
    if err != nil {
        logWarning.Printf("Could not draw text '%s' because: %v", text, err)
        return nil, 0, 0, err
    }

    return textImg, rew + breathe, reh + breathe, nil
}


