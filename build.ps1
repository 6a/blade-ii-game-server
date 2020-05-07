$previous_env = $env:GOOS
$env:GOOS = "linux"

go build -o ./build/b2gameserver

$env:GOOS = $previous_env