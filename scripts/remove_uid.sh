#!/bin/bash

TARGET_DIR="/c/bgg/postapocgame/client"   # 你可以改成你想扫描的目录，例如 ./Client/Assets/

echo "Scanning directory: $TARGET_DIR"

find "$TARGET_DIR" -type f -name "*.cs.uid" | while read file; do
    echo "Removing from git index: $file"
    git rm --cached "$file"
done

echo "Done."
