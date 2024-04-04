import base64
import json
import os
import shutil
import sqlite3
from datetime import datetime, timedelta

import win32crypt  # pip install pypiwin32
from Crypto.Cipher import AES  # pip install pycryptodome

# This is Windows version
# You may want to edit these values to match your OS
local_state_path = os.path.join(os.environ["USERPROFILE"], "AppData", "Local", "Google", "Chrome", "User Data",
                                "Local State")
cookies_path = os.path.join(os.environ["USERPROFILE"], "AppData", "Local", "Google", "Chrome", "User Data", "Default",
                            "Network", "Cookies")


def convert_to_unix_time(expires_utc):
    """
    :param expires_utc: microseconds diff since January, 1601
    :return: epoch time in seconds
    """
    if expires_utc == 0:
        return 0
    # Define the start date as January 1, 1601
    start_date = datetime(1601, 1, 1)
    # Calculate the timedelta from the start date to the expires_utc date
    expires_date = start_date + timedelta(microseconds=expires_utc)
    # Define the Unix epoch start date as January 1, 1970
    epoch_start_date = datetime(1970, 1, 1)
    # Calculate the timedelta from the Unix epoch start date to the expires_utc date
    unix_time = (expires_date - epoch_start_date).total_seconds()
    return unix_time


def convert_samesite(samesite):
    samesite_mapping = {
        -1: "unspecified",
        0: "no_restriction",
        1: "lax",
        2: "strict",
        3: "none",
    }
    return samesite_mapping.get(samesite, "unspecified")


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
        SELECT host_key, name, value, encrypted_value, path, expires_utc, is_secure, is_httponly, has_expires, 
        is_persistent, samesite FROM cookies""")

    encryption_key = get_encryption_key()

    cookies_list = []
    for i, (host_key, name, value, encrypted_value, path, expires_utc, is_secure, is_httponly, has_expires,
            is_persistent, samesite) in enumerate(cursor.fetchall()):
        if not value:
            decrypted_value = decrypt_data(encrypted_value, encryption_key)
        else:
            # already decrypted
            decrypted_value = value

        cookies_list.append({
            "domain": host_key,
            "expirationDate": convert_to_unix_time(expires_utc),
            "hostOnly": False,  # This information is not available in the SQLite database
            "httpOnly": bool(is_httponly),
            "name": name,
            "path": path,
            "sameSite": convert_samesite(samesite),
            "secure": bool(is_secure),
            "session": not bool(is_persistent),
            "storeId": "0",  # This information is not available in the SQLite database
            "value": decrypted_value,
            "id": i + 1
        })

    # commit changes
    db.commit()
    # close connection
    db.close()

    with open('cookies.json', 'w') as f:
        json.dump(cookies_list, f, indent=4)


if __name__ == "__main__":
    main()
