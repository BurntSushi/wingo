BD=bindata
D=data
BINDATA=$(BD)/wingo.png.go \
				$(BD)/close.png.go $(BD)/maximize.png.go $(BD)/minimize.png.go \
				$(BD)/RobotoRegular.ttf.go

all: $(BINDATA)

$(BD)/%.png.go: $(D)/%.png
	go-bindata -f `python2 -c 'print "$*".title()'`Png \
		-i $(D)/$*.png -o $(BD)/$*.png.go -p bindata

$(BD)/%.ttf.go: $(D)/%.ttf
	go-bindata -f `python2 -c 'print "$*".title()'`Ttf \
		-i $(D)/$*.ttf -o $(BD)/$*.ttf.go -p bindata

loc:
	find ./ -name '*.go' -and -not -wholename './tests*' -and -not -wholename './bindata*' -print | sort | xargs wc -l

tags:
	find ./ \( -name '*.go' -and -not -wholename './tests/*' \) -print0 | xargs -0 gotags > TAGS

