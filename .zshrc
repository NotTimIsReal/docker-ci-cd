fucntion compile(){
    echo "Compiling..."
    go build -o .bin/cd-ci
    echo "Compiled"
}
function addPATH(){
    echo "Adding PATH..."
    export PATH=$PATH:$(pwd)/.bin
    echo "Added"
}