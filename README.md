# goload

parse default value from struct tag.

```
type Redis struct {
	Host   string `json:"host" default:"127.0.0.1"`
	Port   int    `json:"port" default:"5678"`
	DB     int8   `json:"DB" default:"5"`
	Enable bool   `json:"enable" default:"true"`
}
```

```

func TestRedis(t *testing.T) {
	c := conf.Redis{DB: 4}
	if err := LoadStruct(&c, "default"); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", c)
}

```

print result is 

```
{  
   Host:127.0.0.1
   Port:5678 
   DB:4
   Enable:true
 }
```
