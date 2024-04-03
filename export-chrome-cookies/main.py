import base64
import json
import os
import shutil
import sqlite3
from datetime import datetime, timedelta

import win32crypt  # pip install pypiwin32
from Crypto.Cipher import AES  # pip install pycryptodome

# This is Windows version
# You may need to edit these values to match your OS
local_state_path = os.path.join(os.environ["USERPROFILE"], "AppData", "Local", "Google", "Chrome", "User Data",
                                "Local State")
cookies_path = os.path.join(os.environ["USERPROFILE"], "AppData", "Local", "Google", "Chrome", "User Data", "Default",
                            "Network", "Cookies")


def get_chrome_datetime(chromedate):
    """Return a `datetime.datetime` object from a chrome format datetime
    Since `chromedate` is formatted as the number of microseconds since January, 1601"""
    if chromedate != 86400000000 and chromedate:
        try:
            return datetime(1601, 1, 1) + timedelta(microseconds=chromedate)
        except Exception as e:
            print(f"Error: {e}, chromedate: {chromedate}")
            return chromedate
    else:
        return ""


def get_encryption_key():
    working_local_state = "Local State"
    if not os.path.isfile(working_local_state):
        shutil.copyfile(local_state_path, working_local_state)
    with open(working_local_state, "r", encoding="utf-8") as f:
        local_state = f.read()
        local_state = json.loads(local_state)

    # decode the encryption key from Base64
    key = base64.b64decode(local_state["os_crypt"]["encrypted_key"])
    # remove 'DPAPI' str
    key = key[5:]
    # return decrypted key that was originally encrypted
    # using a session key derived from current user's logon credentials
    # doc: http://timgolden.me.uk/pywin32-docs/win32crypt.html
    return win32crypt.CryptUnprotectData(key, None, None, None, 0)[1]


def decrypt_data(data, key):
    try:
        # get the initialization vector
        iv = data[3:15]
        data = data[15:]
        # generate cipher
        cipher = AES.new(key, AES.MODE_GCM, iv)
        # decrypt password
        return cipher.decrypt(data)[:-16].decode()
    except:
        try:
            return str(win32crypt.CryptUnprotectData(data, None, None, None, 0)[1])
        except:
            # not supported
            return ""


def main():
    # local sqlite Chrome cookie database path
    # copy the file to current directory
    # as the database will be locked if chrome is currently open
    working_cookies = "Cookies"
    if not os.path.isfile(working_cookies):
        shutil.copyfile(cookies_path, working_cookies)
    # connect to the database
    db = sqlite3.connect(working_cookies)
    # ignore decoding errors
    db.text_factory = lambda b: b.decode(errors="ignore")
    cursor = db.cursor()
    # get the cookies from `cookies` table
    cursor.execute("""
    SELECT host_key, name, value, creation_utc, last_access_utc, expires_utc, encrypted_value 
    FROM cookies""")
    # you can also search by domain, e.g thepythoncode.com
    # cursor.execute("""
    # SELECT host_key, name, value, creation_utc, last_access_utc, expires_utc, encrypted_value
    # FROM cookies
    # WHERE host_key like '%thepythoncode.com%'""")
    # get the AES key
    key = get_encryption_key()
    # dictionary to store cookies
    cookies_dict = {}
    for host_key, name, value, creation_utc, last_access_utc, expires_utc, encrypted_value in cursor.fetchall():
        if not value:
            decrypted_value = decrypt_data(encrypted_value, key)
        else:
            # already decrypted
            decrypted_value = value
        if host_key not in cookies_dict:
            cookies_dict[host_key] = []
        cookies_dict[host_key].append({
            "Cookie name": name,
            "Cookie value (decrypted)": decrypted_value,
            "Creation datetime (UTC)": str(get_chrome_datetime(creation_utc)),
            "Last access datetime (UTC)": str(get_chrome_datetime(last_access_utc)),
            "Expires datetime (UTC)": str(get_chrome_datetime(expires_utc))
        })
        # update the cookies table with the decrypted value
        # and make session cookie persistent
        cursor.execute("""
        UPDATE cookies SET value = ?, has_expires = 1, expires_utc = 99999999999999999, is_persistent = 1, is_secure = 0
        WHERE host_key = ?
        AND name = ?""", (decrypted_value, host_key, name))
    # commit changes
    db.commit()
    # close connection
    db.close()
    with open('cookies.json', 'w') as f:
        json.dump(cookies_dict, f, indent=4)


if __name__ == "__main__":
    main()
