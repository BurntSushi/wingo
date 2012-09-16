BD=bindata
D=data
BINDATA=$(BD)/wingo.png.go \
				$(BD)/close.png.go $(BD)/maximize.png.go $(BD)/minimize.png.go \
				$(BD)/DejaVuSans.ttf.go $(BD)/FreeMono.ttf.go

install: bindata
	go install -p 6 . ./bindata ./cmdusage ./cursors ./focus \
		./frame ./heads ./layout ./logger ./misc ./prompt ./render \
		./stack ./text ./wini ./workspace

git-hooks:
	cp git-hook-pre-commit ./.git/hooks/pre-commit

gofmt:
	gofmt -w *.go cmdusage/*.go cursors/*.go focus/*.go frame/*.go \
		heads/*.go layout/*.go logger/*.go misc/*.go prompt/*.go render/*.go \
		stack/*.go text/*.go wingo-cmd/*.go wini/*.go workspace/*.go
	colcheck *.go */*.go

cmd:
	go install github.com/BurntSushi/wingo/wingo-cmd

bindata: $(BINDATA)

$(BD)/%.png.go: $(D)/%.png
	go-bindata -f `python2 -c 'print "$*".title()'`Png \
		-i $(D)/$*.png -o $(BD)/$*.png.go -p bindata
	gofmt -w $(BD)/$*.png.go

$(BD)/%.ttf.go: $(D)/%.ttf
	go-bindata -f `python2 -c 'print "$*".title()'`Ttf \
		-i $(D)/$*.ttf -o $(BD)/$*.ttf.go -p bindata
	gofmt -w $(BD)/$*.ttf.go

loc:
	find ./ -name '*.go' -and -not -wholename './tests*' -and -not -wholename './bindata*' -print | sort | xargs wc -l

tags:
	find ./ \( -name '*.go' \
					   -and -not -wholename './tests/*' \
						 -and -not -wholename './bindata/*' \) \
			 -print0 \
	| xargs -0 gotags > TAGS

