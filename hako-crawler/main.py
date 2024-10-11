import re
from urllib.request import Request, urlopen
from bs4 import BeautifulSoup

DEBUG = False


def get_from_url(url):
    request_site = Request(url, headers={"User-Agent": "Mozilla/5.0"})
    webpage = urlopen(request_site).read()
    return webpage.decode("utf-8")


def write_text_to_file(text, filename):
    filename = re.sub(r"[^\w_. -]", "_", filename)
    f = open("data/{}".format(filename), "w", encoding="utf-8")
    f.write(text)
    f.close()


def read_text_from_file(filename):
    f = open("data/{}".format(filename), "r", encoding="utf-8")
    text = f.read()
    f.close()
    return text


if DEBUG:
    html = read_text_from_file("_.txt")
else:
    html = get_from_url("https://ln.hako.vn/sang-tac/8476-kiep-nay-la-anh-trai-cua-nhan-vat-chinh")
    # write_text_to_file(html, "_.txt")

soup = BeautifulSoup(html, "html.parser")
chapters = soup.find_all("div", {"class": "chapter-name"})
"""
you may get `HTTP Error 429: Too Many Requests`
you can try again later, and skip downloaded chapters
example skip 50 first chapters like: `for chapter in chapters[50:]`
"""
for chapter in chapters:
    children = chapter.find_all("a", recursive=False)
    child = children[0]
    chapterTitle = child.attrs["title"]
    chapterUrl = "https://ln.hako.vn" + child.attrs["href"]
    chapterHtml = get_from_url(chapterUrl)
    chapterSoup = BeautifulSoup(chapterHtml, "html.parser")
    chapterContent = chapterSoup.find("div", {"id": "chapter-content"})
    chapterData = ""
    if chapterContent is not None:
        paragraphs = chapterContent.find_all("p", id=lambda x: x and x.isdigit())
        for paragraph in paragraphs:
            chapterData += paragraph.text + "\n"
    write_text_to_file(chapterData, chapterTitle + ".txt")
