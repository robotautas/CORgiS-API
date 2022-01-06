import os

os.chdir("../")
print(os.listdir())

counter = 0
for file in os.listdir():
    if file.endswith('go'):
        with open(file, 'r') as f:
            lines = f.readlines()
            print(lines, '\n\n')
            for line in lines:
                counter += 1

print(counter)