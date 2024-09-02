# Multi Request Support

Request Name: posts

    - posts.post.yaml
    - posts.get.yaml
    - posts.put.yaml
    - posts.delete.yaml
    - posts.patch.yaml

Commands:

    - restler get posts
    - restler post posts
    - restler put posts
    - restler delete posts
    - restler patch posts
    - restler all posts?? run all requests in the folder
        - restler all posts --order=post,get,put,patch,delete
        - restler all posts (what will be default order??)

Output files:

    - requests/posts/.get.res.txt
    - requests/posts/.post.res.txt
    - requests/posts/.put.res.txt
    - requests/posts/.delete.res.txt
    - requests/posts/.patch.res.txt
