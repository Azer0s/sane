all:
	@$(MAKE) clear
	@$(MAKE) sane

clear:
	@echo
	@echo "\033[4m\033[1mClearing output folder\033[0m"
	@echo
	@rm -rf bin/ || true
	@mkdir bin

sane:
	@echo
	@echo "\033[4m\033[1mBuilding sane\033[0m"
	@echo
	@go build -o bin/sane main.go

run:
	./bin/sane