package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
}

type Entity struct {
	Id       int
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
}

type Controller struct {
	service *Service
}

func (c *Controller) Registration(writer http.ResponseWriter, request *http.Request) {
	dto := &User{}
	err := json.NewDecoder(request.Body).Decode(dto)
	if err != nil {
		log.Println(err)
		json.NewEncoder(writer).Encode(err)
		return
	}
	err = c.service.Registration(dto)
	if err != nil {
		log.Println(err)
		json.NewEncoder(writer).Encode(err)
		return
	}
	writer.WriteHeader(http.StatusBadRequest)
}

func (c *Controller) List(writer http.ResponseWriter, request *http.Request) {
	users, err := c.service.List()
	if err != nil {
		log.Println(err)
		json.NewEncoder(writer).Encode(err)
		return
	}
	json.NewEncoder(writer).Encode(users)
}

type Service struct {
	repository *Repository
}

func (s *Service) Registration(dto *User) error {
	if dto.Age < 18 {
		return errors.New("age less 18")
	}
	return s.repository.Registration(dto)
}

func (s *Service) List() ([]*User, error) {
	return s.repository.List()
}

type Repository struct {
	db *DAO
}

func (r *Repository) Registration(dto *User) error {
	entity := convertDTOToEntity(dto)
	return r.db.Add(entity)
}

func (r *Repository) List() ([]*User, error) {
	entities, err := r.db.List()
	if err != nil {
		return nil, err
	}
	return convertEntitiesToDTOs(entities), nil
}

type DAO struct {
	sql *sqlx.DB
}

func (d *DAO) Add(entity *Entity) error {
	_, err := d.sql.Exec("INSERT INTO users(email, password, name, age) VALUES (?, ?, ?, ?)", entity.Email, entity.Password, entity.Name, entity.Age)
	return err
}

func (d *DAO) List() ([]*Entity, error) {
	var res []*Entity
	err := d.sql.Select(&res, "SELECT * FROM users")
	return res, err
}

func main() {
	controller := createController()
	route := createRoute(controller)
	log.Fatal(http.ListenAndServe(":8080", route))
}

func createController() *Controller {
	return &Controller{service: createService()}
}

func createService() *Service {
	return &Service{repository: createRepository()}
}

func createRepository() *Repository {
	return &Repository{db: createDAO()}
}

func createDAO() *DAO {
	sql := connectSQL()
	migrate(sql)
	return &DAO{sql: sqlx.NewDb(sql, "sqlite")}
}

func migrate(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users (email VARCHAR(30) PRIMARY KEY, password VARCHAR(30), name VARCHAR(30), age INTEGER)")
	if err != nil {
		panic(err)
	}
}

func connectSQL() *sql.DB {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	return db
}

func createRoute(controller *Controller) *chi.Mux {
	route := chi.NewRouter()
	route.Post("/registration", controller.Registration)
	route.Get("/list", controller.List)
	return route
}

func convertDTOToEntity(dto *User) *Entity {
	return &Entity{
		Id:       0,
		Email:    dto.Email,
		Password: dto.Password,
		Name:     dto.Name,
		Age:      dto.Age,
	}
}

func convertEntitiesToDTOs(entities []*Entity) []*User {
	users := make([]*User, len(entities))
	for i := range entities {
		users[i] = convertEntityToDTO(entities[i])
	}
	return users
}

func convertEntityToDTO(entity *Entity) *User {
	return &User{
		Email:    entity.Email,
		Password: entity.Password,
		Name:     entity.Name,
		Age:      entity.Age,
	}
}
