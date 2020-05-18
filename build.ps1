$previous_env = $env:GOOS
$env:GOOS = "linux"

go build -o ./build/gameserver

$env:GOOS = $previous_env