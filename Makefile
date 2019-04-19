build:
	time go install

install: build
	rsync -avPz static templates /srv/tv/
