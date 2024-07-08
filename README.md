# You ever wondered how Git works ???

Neither did I until I really started to wonder how does version control work. It truly is a simple and effective way of doing things under the hood


## An Attempt

What my program tries to do is represent what goes on behind the scenes with Git with different commands. Which I will document below


## Commands

### init 
This basically as it suggest initialises what we know as the .git folder. Within this folder we create the some other important folders such as the objects folder and the refs folder
Keep the objects folder in mind as this will the main focus and the bread and butter for Git.

### hash-object -w [file-name]
This command allows you to creat a blob object in the objects folder. But why not just store the data plainly?? Cause that is never a good idea plus
how do you think Git knows the file has changed? Yeah, it compares these hashes to see if it has changed making it easier to identify. So what Hashing Algo does git use? 
In this program and in Git currently SHA-1. This can change in the future to SHA-256 but is not necessary due to how Git uses hashing.

### cat-file [object-hash]
This command allows you to read the blob object we created if we ran the command above. So you enter the hash you produced and this gives you a read out of the file

### write-tree
This is the second of tree (pun intended) objects that Git stores. Git uses this object to store basically your file structure. So how it is store is based on the entry. 

This is an example of what the object looks like before it is hashed (there is usually a null value between the entry name and the hash value)
  040000 dir_name1 <tree_sha_1>
  040000 dir_name2 <tree_sha_2>
  100644 file_name1 <blob_sha_1>

The output of this command is the hash of the above content

### ls-tree
This allows you to read the hash of a tree object 



I only mentioned 2 of 3 objects that Git uses. The final object type it a commit type. So this is when you normally git commit an object and it stores it. I'm getting to this one though.
