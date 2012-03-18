loc:
	find ./ -name '*.go' -and -not -wholename './tests*' -print | sort | xargs wc -l

tags:
	find ./ \( -name '*.go' -and -not -wholename './tests/*' \) -print0 | xargs -0 gotags > TAGS

