## Flow WEBSOCKET

1. Login
	url : http://localhost:3000/login
	metode : POST
	payload :
	{
		"username":"ical",
		"password":"admin123"
	}

	respon :
	{
		"token":"12345"
	}

	flow :
	token pakai expired 2 detik

2. masuk WS
	ws://localhost:3000/ws

	message : 
	{
        "topic":"login",
        "proc":"",
        "data":"EGJp4gVNAaJDa8zyf2vIQyjvBZoXNVXeRsuPlo1rffWyoLmDOezVjF98hSvHjU5HW1alTBvPTE5fKYrjMToU88LHk/k0eOOF38V48IgsjTEDYYQxm4e0e4lQG2xjbPwdP+56wg==",
        "from":"",
        "exp":0
    }
	flow :
	Jika token sudah waktunya kurang dari saat ini maka tidak bisa masuk

3. message WS

    Edit :
    ```
	{"topic":"user","proc":"edit","data":"{\"username\":\"ical\",\"fullname\":\"Fahrizal\",\"email\":\"fahrizal.khoirianto@gmail.com\",\"companyId\":\"4000\"}"}
    ```

    Create :
    ```
    {"topic":"user","proc":"new","data":"{\"fullname\":\"Radiyah Salma\",\"email\":\"radiyah.salma@gmail.com\",\"hp\":\"+6287812730567\",\"username\":\"salma\",\"password\":\"salmazxy\",\"companyId\":\"43\",\"roleId\":\"0\",\"officeId\":\"0\",\"departementId\":\"0\"}"}
    ```

    Delete :
    ```
    {"topic":"user","proc":"delete","data":"{\"username\":\"ahmad\"}"}
    ```    

    Search :
    ```
    {"topic":"user","proc":"search","data":"{\"username\":\"ical\",\"status\":\"1\"}"}
    ```