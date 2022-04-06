# QOR Admin

Instantly create a beautiful, cross platform, configurable Admin Interface and API for managing your data in minutes.

[![GoDoc](https://godoc.org/github.com/qor/admin?status.svg)](https://godoc.org/github.com/qor/admin)
[![Build Status](https://travis-ci.com/qor/admin.svg?branch=master)](https://travis-ci.com/qor/admin)

**For security issues, please send us an email to security@getqor.com and give us time to respond BEFORE posting as an issue or reporting on public forums.**

## Features

- Generate Admin Interface for managing data
- RESTFul JSON API
- Association handling
- Search and filtering
- Actions/Batch Actions
- Authentication and Authorization
- Extendability

## Quick Start

```go
package main

import (
  "fmt"
  "net/http"
  "github.com/jinzhu/gorm"
  _ "github.com/mattn/go-sqlite3"
  "github.com/qor/admin"
)

// Create a GORM-backend model
type User struct {
  gorm.Model
  Name string
}

// Create another GORM-backend model
type Product struct {
  gorm.Model
  Name        string
  Description string
}

func main() {
  DB, _ := gorm.Open("sqlite3", "demo.db")
  DB.AutoMigrate(&User{}, &Product{})

  // Initialize
  Admin := admin.New(&admin.AdminConfig{DB: DB})

  // Allow to use Admin to manage User, Product
  Admin.AddResource(&User{})
  Admin.AddResource(&Product{})

  // initialize an HTTP request multiplexer
  mux := http.NewServeMux()

  // Mount admin interface to mux
  Admin.MountTo("/admin", mux)

  fmt.Println("Listening on: 9000")
  http.ListenAndServe(":9000", mux)
}
```

`go run main.go` and visit `localhost:9000/admin` to see the result!

## How to use remoteSelector with publish2.version integrated record

### **For has many relationship**
Suppose we have 2 models Factory and Item. Factory **has many** Items.

In the struct, you need add a field `resource.CompositePrimaryKeyField` to the "many" side, which is `Item` here.
```go
type Factory struct {
	gorm.Model
	Name string

	publish2.Version
	Items       []Item `gorm:"many2many:factory_items;association_autoupdate:false"`
	ItemsSorter sorting.SortableCollection
}

type Item struct {
	gorm.Model
	Name string
	publish2.Version

	// github.com/qor/qor/resource
	resource.CompositePrimaryKeyField // Required
}
```

Then define a remote resource selector. You need configure the `ID` meta like below to make it support composite primary key, this is mandatory.
```go
func generateRemoteItemSelector(adm *admin.Admin) (res *admin.Resource) {
	res = adm.AddResource(&Item{}, &admin.Config{Name: "ItemSelector"})
	res.IndexAttrs("ID", "Name")

	// Required. Convert single ID into composite primary key
	res.Meta(&admin.Meta{
	Name: "ID",
	Valuer: func(value interface{}, ctx *qor.Context) interface{} {
		if r, ok := value.(*Item); ok {
			// github.com/qor/qor/resource
			return resource.GenCompositePrimaryKey(r.ID, r.GetVersionName())
		}
		return ""
	},
	})

	return res
}
```

Last, use it in the Factory resource.
```go
itemSelector := generateRemoteItemSelector(adm)
factoryRes.Meta(&admin.Meta{
	Name: "Items",
	Config: &admin.SelectManyConfig{
	RemoteDataResource: itemSelector,
	},
})
```

### **For has one relationship**
Suppose we have 2 models. Factory and Manager. Factory **has one** Manager.

First, In the struct, you need add a field `resource.CompositePrimaryKeyField` to the "one" side, which is `Manager` here.
```go
type Factory struct {
	gorm.Model
	Name string
	publish2.Version

	ManagerID          uint
	ManagerVersionName string // Required. in "xxxVersionName" format.
	Manager            Manager
}

type Manager struct {
	gorm.Model
	Name string
	publish2.Version

	// github.com/qor/qor/resource
	resource.CompositePrimaryKeyField // Required
}
```

Then define a remote resource selector. You need configure the `ID` meta like below to make it support composite primary key, this is mandatory.
```go
func generateRemoteManagerSelector(adm *admin.Admin) (res *admin.Resource) {
	res = adm.AddResource(&Manager{}, &admin.Config{Name: "ManagerSelector"})
	res.IndexAttrs("ID", "Name")

	// Required. Convert single ID into composite primary key
	res.Meta(&admin.Meta{
		Name: "ID",
		Valuer: func(value interface{}, ctx *qor.Context) interface{} {
			if r, ok := value.(*Manager); ok {
				// github.com/qor/qor/resource
				return resource.GenCompositePrimaryKey(r.ID, r.GetVersionName())
			}
			return ""
		},
	})

	return res
}

Last, use it in the Factory resource.
```go
managerSelector := generateRemoteManagerSelector(adm)
factoryRes.Meta(&admin.Meta{
	Name: "Manager",
	Config: &admin.SelectOneConfig{
		RemoteDataResource: managerSelector,
	},
})
```

If you need to overwrite Collection. you have to pass composite primary key as the first element of the returning array instead of ID.
```go
factoryRes.Meta(&admin.Meta{
  Name: "Items",
  Config: &admin.SelectManyConfig{
	Collection: func(value interface{}, ctx *qor.Context) (results [][]string) {
		if c, ok := value.(*Factory); ok {
		var items []Item
		ctx.GetDB().Model(c).Related(&items, "Items")

		for _, p := range items {
		// The first element must be the composite primary key instead of ID
		results = append(results, []string{resource.GenCompositePrimaryKey(p.ID, p.GetVersionName()), p.Name})
		}
		}
		return
	},
	RemoteDataResource: itemSelector,
  },
})
```

## To support assign associations when creating a new version
If you want to assign associations when creating a new version of object immediately. You need to define a function called `AssignVersionName` to the versioned struct with **pointer** receiver which should contains the generating new version name's logic and assign the new version name to the object.
e.g.
```go
func (fac *Factory) AssignVersionName(db *gorm.DB) {
	var count int
	name := time.Now().Format("2006-01-02")
	if err := db.Model(&CollectionWithVersion{}).Where("id = ? AND version_name like ?", fac.ID, name+"%").Count(&count).Error; err != nil {
    panic(err)
  }
	fac.VersionName = fmt.Sprintf("%s-v%v", name, count+1)
}
```


## Live DEMO

* Live Demo [http://demo.getqor.com/admin](http://demo.getqor.com/admin)
* Source Code of Live Demo [https://github.com/qor/qor-example](https://github.com/qor/qor-example)

## Documentation

<https://doc.getqor.com/admin>

To print all registered routes
```go
// adm is a QOR admin instance
adm.GetRouter().PrintRoutes()
```

## ViewPath Note

QOR was developed before go mod was introduced. So it still support go path while finding its template files. The priority is

1. check vendor, if not found
2. check $GOPATH/pkg/mod/github.com/qor/admin@v0.x/views. the version would be detected automatically by your go.mod file, if still not found
3. load view path from $GOPATH/src/github.com/qor/admin/views


So if you want to use the template under the pkg/mod, make sure $GOPATH/src/github.com/qor/admin is absent.

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
