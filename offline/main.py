import data_process
import md5
import sender

from config import LABEL_FILENAME, LABEL_FILEMD5

if __name__ == "__main__":
    files_dict = data_process.run()

    print(files_dict)
    for file in files_dict:
        # files_dict {"pod_info": {"name": "2020-02-12-12-12-12-pod_info.csv", "md5": "1231245345"}}
        res = md5.get_file_md5(files_dict[file][LABEL_FILENAME])
        files_dict[file][LABEL_FILEMD5] = res

    sender.run(files_dict)
