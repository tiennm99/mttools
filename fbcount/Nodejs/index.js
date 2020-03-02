const fs = require("fs");
const login = require("facebook-chat-api");
const readline = require("readline");
const readlineSync = require('readline-sync');
 
var email = readlineSync.question('Please enter your email or phone: ');
var password = readlineSync.question('Please enter your password: ');

var rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

const obj = {email, password};
login(obj, (err, api) => {
    if(err) {
        switch (err.error) {
            case 'login-approval':
                console.log('Enter code > ');
                rl.on('line', (line) => {
                    err.continue(line);
                    rl.close();
                });
                break;
            default:
                console.error(err);
        }
        return;
    }
    console.log("get thread list...");
    api.getThreadList(1000, null, [], (err, list) => {
        if (err) throw err;
        if (list) {
            console.log("sort...");
            list.sort(function(a, b) {
                return a.messageCount - b.messageCount;
            })
            console.log("write...");
            list.forEach(function(value, i) {
                fs.appendFile('output.txt', (i + 1).toString() + "\t" + value.name + "\t" + value.threadID + "\t" + value.messageCount, (err) => {
                    if (err) throw err;
                })
            });
            console.log("done!");
        }
    });
});