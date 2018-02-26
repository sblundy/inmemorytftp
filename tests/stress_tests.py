#!/usr/bin/env python3

import tftpy
import cStringIO
import threading
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('-n', dest='hostname', metavar='Hostname to connect to', default='localhost')
parser.add_argument('-p', dest='port', metavar='Port to connect to', type=int, default=69)
parser.add_argument('-t', dest='num_threads', metavar='Number of threads', type=int, default=4)
parser.add_argument('-s', dest='file_size', metavar='Size of files to send', type=int, default=512)
parser.add_argument('-f', dest='num_files', metavar='Number of times to send each file by each threads', type=int, default=1000)
args = parser.parse_args()

class Tester(threading.Thread):
    def run(self) :
        client = tftpy.TftpClient(args.hostname, args.port)
        contents = ''

        for x in range(0, args.file_size / 10):
            contents = contents + "1234567890"

        if args.file_size % 10 > 0:
            contents = contents + "1234567890"[:args.file_size % 10]

        for x in range(0, args.num_files):
            file = cStringIO.StringIO(contents)
            output = cStringIO.StringIO()
            filename = 'test' + str(x) + '.txt'
            client.upload(filename, file)
            client.download(filename, output)

testers = []
for x in range(0, args.num_threads):
    tester = Tester()
    tester.start()
    testers.append(tester)


for tester in testers:
    tester.join()