genProtoCS(){
  dir=$(pwd)
  echo "proto dir: $dir"
  proDir=$(dirname $dir)
  echo "project Dir: $proDir"

  # 输出目录（你可以改成你的 C# 工程目录）
  outDir="${proDir}/client/Scripts/Protocol"

  mkdir -p $outDir

  # 生成 C#
  $dir/protoc.exe \
      -I=$dir/csproto \
      --csharp_out=$outDir \
      $dir/csproto/*.proto

  echo "gen csharp proto done"

  # 可选：格式化 C#（如果你有 dotnet-format）
  # dotnet format $outDir
}

genProtoCS
