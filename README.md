# Tokype programming language



Creator: Mr. Artem.
The license is located in the same folder as the executable file. It is recommended to review the license before use.



## Tokype syntax help:



### Creating variables:

```tokype
a = 10
b = "Hello"
c = true
d = [1, 2, 3]
e = {"Hello" :: "world", 1 :: 2}
```
### Built-in functions:
```tokype
print(10)

input("You can enter text: ")

a = [10, 20, 30]
b = len(a)

print(b)

print(negate(5))

push(a, 40)
print(a)

print(pop(a))
print(first(a))
print(rest(a))
print(contains(a, 10))

c = get_time()
print(c)

reserve(a, 10)
big_list = repeat(5, 10)

print("List: " + big_list.[1])
```
### Functions:
```tokype
funct main():
   result = 10
   print(result)
end

main()
```
### Conditions:
```tokype
if 5 > 7:
   print("5 > 7")
elif 5 > 6:
   print("5 > 6")
else:
   print("5 < 7 and 5 < 6")
end
```
### Loops:
```tokype
for i = 1, 10:
   sum = sum + i
end

x = 10

while x > 0:
   print(x)
   x = x - 1
end
```
### Lists and dictionaries:
```tokype
a = [10, 20, 30]
print(a.[1])

b = {10 :: "10", "20" :: "20"}
print(b.["20"])
```
### Operators:
```tokype
a = 10 + 5
b = 10 - 5
c = 10 * 5
d = 10 / 5
e = 10 % 3

print(5 == 5)
print(5 ?= 5)
print(5 < 10)
print(5 > 10)
print(5 <= 5)
print(5 >= 5)

print("Hello, " + "world!")
```
### Comments:
```tokype
~~comment
```
## CLI commands:

### Version:
```bash
tokype version
```
### License:
```bash
tokype license
```
### Help:
```bash
tokype help
```
### Run:
```bash
tokype <file>.tp
```
Good luck :)
