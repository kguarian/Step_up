import sys

print("hello world")
x=iter(sys.stdout.readline, "")

for l in x :
    print(l)