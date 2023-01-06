package controllers

import (
	"dolphin/app/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func FindUser(c *gin.Context) {
	var result map[string]interface{}

	// map query result to var result
	err := models.DB.Table("users").Where("id = ?", c.Param("id")).Find(&result).Error

	if len(result) == 0 {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// build transformer from json file
	jsonFile, err := os.Open("transformers/user/result.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var transformer map[string]interface{}
	json.Unmarshal([]byte(byteValue), &transformer)

	// remap result to transformer
	for key := range transformer {
		//ini mesti kasih pengecekan apakah dia object atau array atau value
		transformer[key] = result[key]
	}

	// return transformed result
	c.JSON(http.StatusOK, transformer)
}

func FindUsers(c *gin.Context) {
	var results []map[string]interface{}
	var collection []map[string]interface{}

	// map query results to var results
	err := models.DB.Table("users").Find(&results).Error

	if len(results) == 0 {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// build transformer from json file
	jsonFile, err := os.Open("transformers/user/collection.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var transformer map[string]interface{}
	json.Unmarshal([]byte(byteValue), &transformer)

	// remap results to collection
	for keyUser, result := range results {
		collection = append(collection, map[string]interface{}{})

		// remap result to transformer
		for key := range transformer {
			collection[keyUser][key] = result[key]
		}
	}

	// return collection of transformed result
	c.JSON(http.StatusOK, collection)
}

func CreateUser(c *gin.Context) {
	// build transformer from json file
	jsonFile, err := os.Open("transformers/user/insertInput.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var input map[string]interface{}
	json.Unmarshal([]byte(byteValue), &input)

	// Validate input
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Table("users").Create(input)
}

func UpdateUser(c *gin.Context) {
	// build transformer from json file
	jsonFile, err := os.Open("transformers/user/updateInput.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var input map[string]interface{}
	json.Unmarshal([]byte(byteValue), &input)

	// Validate input
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updateResult = models.DB.Table("users").Where("id = ?", c.Param("id")).Updates(input)

	if updateResult.Error != nil {
		fmt.Println(updateResult.Error)
		c.JSON(http.StatusBadRequest, gin.H{"error": updateResult.Error})
		return
	}

	// map query result to var result
	// updateResult.RowsAffected always return 0 if row found but no changes
	// meaning, we need to reselect the row to check if data exist or not
	var result map[string]interface{}

	models.DB.Table("users").Where("id = ?", c.Param("id")).Find(&result)

	if len(result) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// build transformer from json file
	jsonFile, err = os.Open("transformers/user/result.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ = ioutil.ReadAll(jsonFile)

	var transformer map[string]interface{}
	json.Unmarshal([]byte(byteValue), &transformer)

	// remap result to transformer
	for key := range transformer {
		//ini mesti kasih pengecekan apakah dia object atau array atau value
		transformer[key] = result[key]
	}

	// return transformed result
	c.JSON(http.StatusOK, transformer)
}

func DeleteUser(c *gin.Context) {}
