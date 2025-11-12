genProto(){
  dir=$(pwd)
  echo "proto dir: $dir"
  proDir=$(dirname $dir)
  echo "project Dir: $proDir"
  $dir/protoc.exe -I=$dir/csproto --go_out=${proDir} $dir/csproto/*.proto
  echo "gen proto done"
  echo "gofmt -w ${proDir}/server"
  gofmt -w ${proDir}/server
}
genProto
