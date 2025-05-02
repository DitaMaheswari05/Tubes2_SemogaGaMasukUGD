module github.com/wiwekaputera/Tubes2_SemogaGaMasukUGD/backend

go 1.18

require github.com/PuerkitoBio/goquery v1.8.0

require (
	github.com/andybalholm/cascadia v1.3.1 // indirect
	golang.org/x/net v0.0.0-20210916014120-12bc252f5db8 // indirect
)

// use the code in the local ./recipeFinder directory instead of pulling it remotely.
replace github.com/wiwekaputera/Tubes2_SemogaGaMasukUGD/backend/recipeFinder => ./recipeFinder
