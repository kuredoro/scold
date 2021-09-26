process.stdin.setEncoding("utf8");

var a;
process.stdin.on("data", (data) => {
    var parts = data.split(' ')
    console.log(Number(parts[0]) + Number(parts[1]))
});
