import sys
import os
import traceback
import main
try:
        import mariadb
except:
        pass

DB_USER = os.getenv('DB_USER', 'user')
DB_PASS = os.getenv('DB_PASS', 'pass')
DB_NAME = os.getenv('DB_NAME', main.BOT_NAME + '_db')
DB_HOST = os.getenv('DB_HOST', '127.0.0.1')
DB_PORT = os.getenv('DB_PORT', '3306')
USE_DB = os.getenv('USE_DB', '0')

DB_VER = 1

CONN = None

def attempt_connect_to_mariadb():
        if USE_DB == '0':
                return False
        try:
                global CONN
                CONN = mariadb.connect(
                        user = DB_USER,
                        password = DB_PASS,
                        host = DB_HOST,
                        port = int(DB_PORT),
                )
                cur = CONN.cursor()
                cur.execute("SHOW DATABASES;")
                dbs = cur.fetchall()
                if (DB_NAME,) not in dbs:
                        initialize_mariadb_database()
                cur.close()
                CONN.close()

                CONN = mariadb.connect(
                        user = DB_USER,
                        password = DB_PASS,
                        host = DB_HOST,
                        port = int(DB_PORT),
                        database = DB_NAME,
                )
        except:
                print(f"UNABLE TO CONNECT TO MARIADB ->{DB_NAME}<-, SKIPPING...")
                return False


        print(f"MARIADB ->{DB_NAME}<- OK!")
        cur = CONN.cursor()

        try:
                initialize_mariadb_table()
                check_mariadb_records()
        except:
                return False

        cur.close()
        return True

def initialize_mariadb_database():
        try:
                cur = CONN.cursor()
                cur.execute(f"CREATE DATABASE {DB_NAME} CHARACTER SET utf8mb4;")
                CONN.commit()
                cur.close()
                print("DATABASE INITIALIZED.")
        except:
                print(traceback.format_exc())
                raise Exception("ERROR CREATING DATABASE.")


def initialize_mariadb_table():
        try:
                cur = CONN.cursor()
                cur.execute(f"SHOW TABLES;")
                tables = cur.fetchall()
                if ('line',) not in tables:
                        cur.execute("CREATE TABLE line (line_id VARCHAR(128), tg_id VARCHAR(128), tg_title VARCHAR(255), line_link VARCHAR(512), auto_emoji BOOL);")
                if ('properties',) not in tables:
                        cur.execute("CREATE TABLE properties (name VARCHAR(128) PRIMARY KEY, value VARCHAR(128));")
                        cur.execute("INSERT properties (name, value) VALUES (?, ?)", ('DB_VER', DB_VER))
                if ('stickers',) not in tables:
                        cur.execute("CREATE TABLE stickers (user_id BIGINT, tg_id VARCHAR(128), tg_title VARCHAR(255), timestamp BIGINT);")
                CONN.commit()
                cur.close()
        except:
                print("ERROR INITIALIZING MARIADB TABLE! EXITING...")
                print(traceback.format_exc())
                sys.exit()


def check_mariadb_records():
        cur = CONN.cursor()
        try:
                cur.execute("SELECT value FROM properties WHERE name=?", ('DB_VER',))
                selected_db_ver = cur.fetchone()[0]
                # Upgrade tables, ONE minor revision.
                if selected_db_ver == '0':
                        cur.execute("ALTER TABLE line DROP PRIMARY KEY;")
                        cur.execute("ALTER TABLE line ADD (line_link VARCHAR(512), auto_emoji BOOL)")
                        cur.execute("UPDATE line SET auto_emoji=? WHERE auto_emoji IS NULL;", (False,))
                        cur.execute("UPDATE properties SET value=? WHERE name=?", (1, 'DB_VER'))
                        print("UPGRADED DATABASE 1 MINOR REVISION.")
        except Exception as e:
                print("ERROR SANITIZING TABLE! CHECK MARIADB NOW! EXITING...\n" + str(e))
                sys.exit()

        CONN.commit()
        cur.close()


def query_line_sticker(line_id: str):
        if CONN is None:
                return None
        try:
                cur = CONN.cursor()
                cur.execute(
                        "SELECT tg_id, auto_emoji FROM line WHERE line_id=?", (line_id,)
                )
                tg_id = cur.fetchall()
                cur.close()
                return tg_id
        except:
                cur.close()
                return None


def insert_line_sticker(line_id, line_link, tg_id, tg_title, is_auto_emoji):
        if CONN is None:
                return
        try:
                cur = CONN.cursor()
                cur.execute(
                        "INSERT line (line_id, line_link, tg_id, tg_title, auto_emoji) VALUES (?, ?, ?, ?, ?)", (line_id, line_link, tg_id, tg_title, is_auto_emoji)
                )
                CONN.commit()
                print(f"INSERT OK -> {line_id} | {tg_id} | {tg_title} | {str(is_auto_emoji)}")
                
        except Exception as e:
                print("FAILED TO INSERT, SKIPPING...\n" + str(e))
        cur.close()


def insert_user_sticker(user_id, tg_id, tg_title, timestamp):
        if CONN is None:
                return
        try:
                cur = CONN.cursor()
                cur.execute(
                        "INSERT stickers (user_id, tg_id, tg_title, timestamp) VALUES (?, ?, ?, ?)", (user_id, tg_id, tg_title, timestamp)
                )
                CONN.commit()
                print(f"INSERT OK -> {user_id} | {tg_id} | {tg_title} | {str(timestamp)}")
                
        except Exception as e:
                print("FAILED TO INSERT, SKIPPING...\n" + str(e))
        cur.close()


def query_user_sticker(user_id):
        if CONN is None:
                return None
        try:
                cur = CONN.cursor()
                cur.execute(
                        "SELECT tg_id, tg_title, timestamp FROM stickers WHERE user_id=?", (user_id,)
                )
                stickers = cur.fetchall()
                cur.close()
                return stickers
        except:
                cur.close()
                return None


def delete_user_sticker(tg_id):
        if CONN is None:
                return None
        try:
                cur = CONN.cursor()
                cur.execute(
                        "DELETE FROM stickers WHERE tg_id=?", (tg_id,)
                )
                CONN.commit()
                cur.close()
        except:
                cur.close()
                return None