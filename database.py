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

DB_VER = 0


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
                        cur.execute("CREATE TABLE line (line_id VARCHAR(128) PRIMARY KEY, tg_id VARCHAR(128), tg_title VARCHAR(255));")
                if ('properties',) not in tables:
                        cur.execute("CREATE TABLE properties (name VARCHAR(128) PRIMARY KEY, value VARCHAR(128));")
                if ('stickers',) not in tables:
                        cur.execute("CREATE TABLE stickers (user_id BIGINT, tg_id VARCHAR(128), tg_title VARCHAR(255), timestamp BIGINT);")
                CONN.commit()
                cur.close()
                print("DATABASE TABLE INITIALIZED.")
        except:
                print("ERROR INITIALIZING MARIADB TABLE! EXITING...")
                print(traceback.format_exc())
                exit(1)


def check_mariadb_records():
        cur = CONN.cursor()
        try:
                cur.execute("SELECT value FROM properties WHERE name=?", ('DB_VER',))
                db_version = cur.fetchone()[0]
                if db_version < DB_VER:
                        #Reserved.
                        pass
        except:
                cur.execute("INSERT INTO properties (name, value) VALUES (?, ?)", ('DB_VER', DB_VER))

        CONN.commit()
        cur.close()



def query_tg_id_by_line_id(line_id: str):
        if CONN is None:
                return None
        try:
                cur = CONN.cursor()
                cur.execute(
                        "SELECT tg_id FROM line WHERE line_id=?", (line_id,)
                )
                tg_id = cur.fetchone()
                cur.close()
                return tg_id[0]
        except:
                cur.close()
                return None


def insert_line_and_tg_id(line_id, tg_id, tg_title):
        if CONN is None:
                return
        try:
                cur = CONN.cursor()
                cur.execute(
                        "INSERT INTO line (line_id, tg_id, tg_title) VALUES (?, ?, ?)", (line_id, tg_id, tg_title)
                )
                CONN.commit()
                print(f"INSERT OK -> {line_id} | {tg_id} | {tg_title}")
                
        except Exception as e:
                print("FAILED TO INSERT, SKIPPING...\n" + str(e))
        cur.close()


def insert_user_sticker(user_id, tg_id, tg_title, timestamp):
        if CONN is None:
                return
        try:
                cur = CONN.cursor()
                cur.execute(
                        "INSERT INTO stickers (user_id, tg_id, tg_title, timestamp) VALUES (?, ?, ?, ?)", (user_id, tg_id, tg_title, timestamp)
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