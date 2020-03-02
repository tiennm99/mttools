import json, os, fbchat
from fbchat.models import *
from fbchat import log, Client

def main():
    file = open("info.txt", "r")
    info = file.readlines()
    un = info[0]
    pw = info[1]
    if os.path.isfile('session.json'):
        with open('session.json') as f:
            session_cookies = json.load(f)
    else:
        session_cookies = None
    client = Client(un, pw, session_cookies=session_cookies)
    session_cookies_new = client.getSession()
    with open('session.json', 'w') as outfile:
        json.dump(session_cookies_new, outfile)
    print("fetch all users...")
    all_users = client.fetchAllUsers()
    print("fectch thread info...")
    detailed_users = [list(client.fetchThreadInfo(user.uid).values())[0] for user in all_users]
    print("sort...")
    sorted_detailed_users = sorted(detailed_users, key=lambda u: u.message_count if u.message_count is not None else -1, reverse=True)
    print("writing...")
    output = open("output.txt", "w+")
    for i,u in enumerate(sorted_detailed_users):
        output.write(str(i + 1) + "\t" + str(u.message_count) + "\t" + u.uid + "\t" + u.name + "\n")
    output.close()
    client.logout()
    print("done!")

if __name__ == '__main__':
    main()