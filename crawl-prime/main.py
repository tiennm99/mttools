import requests
from bs4 import BeautifulSoup

url = "http://compoasso.free.fr/primelistweb/page/prime/liste_online_en.php"
max = 1_000_000
current = 0

while current < max:
  data = {
    "primePageInput": "600",
    "numberInput": current,
  }

  response = requests.post(url, data=data)

  if response.status_code == 200:
    soup = BeautifulSoup(response.text, 'html.parser')
    table = soup.find('table')
    if table:
      form = table.find('form')
      if form:
        table1 = form.find('table')
        if table1:
          td_elements = table1.find_all('td')
          output_file = "primes.txt"

          with open(output_file, "a", encoding="utf-8") as file:
            for td in td_elements[0:-10]:
              file.write(td.text + "\n")

          current = int(td_elements[-11].text) + 1
  else:
    print("Error - current: " + current)
