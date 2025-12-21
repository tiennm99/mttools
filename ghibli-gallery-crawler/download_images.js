const fs = require('fs')
const request = require('request')
const {
    cookieCompare
} = require('tough-cookie')

const download = (url, path, callback) => {
    request.head(url, (err, res, body) => {
        request(url)
            .pipe(fs.createWriteStream(path))
            .on('close', callback)
    })
}

// const url = 'https://…'
// const path = './images/image.png'

// download(url, path, () => {
//   console.log('✅ Done!')
// })

// let urlConst = "https://www.iamag.co/wp-content/uploads/2017/07/yourname-background"
// for (let i = 1; i < 200; i++) {
//     download(urlConst + i + ".jpeg", "img/" + i + ".jpeg", () => {
//         console.log("Done: " + i);
//     })
// }

function lpad(n, width, z) {
    z = z || '0';
    n = n + '';
    return n.length >= width ? n : new Array(width - n.length + 1).join(z) + n;
}

let ghibli = "http://www.ghibli.jp/gallery/";
let listFilm = ["marnie", "kaguyahime", "kazetachinu", "kokurikozaka", "karigurashi", "ponyo", "ged", "chihiro"];
listFilm.forEach(film => {
    let filmDir = "./" + film;
    fs.mkdir(filmDir, function(err) {
        if (err) {
            console.log(err);
        } else {
            console.log("Created " + filmDir);
        }
    });
    for (let i = 1; i <= 50; i++) {
        let imgName = lpad(i, 3) + ".jpg";
        let imgUrl = ghibli + film + imgName;
        console.log(imgUrl);
        download(imgUrl, filmDir + "/" + imgName, () => {
            console.log("Done: " + imgName);
        })
    }
})