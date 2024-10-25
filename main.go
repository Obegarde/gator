package main
import(
	"fmt"
	"github.com/obegarde/gator/internal/config"
	"os"
	"database/sql"
	"github.com/lib/pq"
	"github.com/obegarde/gator/internal/database"
	"context"
	"github.com/google/uuid"	
	"time"
	
)


type state struct{
	config *config.Config
	db *database.Queries
}

type command struct{
	name string
	args []string
}

type commands struct{
	handlers map[string]func(*state, command) error
}


func main(){
	configData, err := config.Read()
	if err != nil{
	fmt.Printf("Read Error: %v\n", err)	
	}
	theState := state{}
	theState.config = &configData
	db, err := sql.Open("postgres",theState.config.DbURL)
	if err != nil{
		fmt.Printf("SqlError: %v\n", err)
	}
	defer db.Close()
	dbQueries := database.New(db)
	theState.db = dbQueries
	theCommands := newCommands()
	theCommands.register("login", handlerLogin)
	theCommands.register("register", handlerRegister)
	theCommands.register("reset",handlerResetUsers)
	if len(os.Args) < 2{
		os.Exit(1)
	}
	
	theCommand := new(command)
	theCommand.name = os.Args[1]
	theCommand.args = os.Args[2:]
	err = theCommands.run(&theState,*theCommand)
	if err != nil{
	fmt.Printf("run Error: %v\n", err)
	os.Exit(1)
	}

}
func handlerResetUsers(s *state, _ command) error{
	fmt.Println("Attempting to reset users")
	err := s.db.ResetUsers(context.Background())
	if err != nil{
	return err
	}
	return nil
}


func handlerRegister(s *state, cmd command) error{
	if len(cmd.args) < 1{
	return fmt.Errorf("Too few arguements, please pass in a name together with the register command")
	}
	newUserParams := new(database.CreateUserParams)
	newUserContext := context.Background()
	newUserParams.ID = uuid.New()
	newUserParams.Name = cmd.args[0]
	newUserParams.CreatedAt = time.Now()
	newUserParams.UpdatedAt = time.Now()
	fmt.Printf("Attempting to create user with name: %s\n", newUserParams.Name)
	newUser,err := s.db.CreateUser(newUserContext, *newUserParams)
	if err != nil{
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505"{
			return fmt.Errorf("user already exists")	
		}
	return err
	}
	err = config.SetUser(newUserParams.Name, *s.config)
	if err != nil{
	return err
	}
	fmt.Printf("User successfully registered\nUserData: %v\n",newUser)
	return nil
}

func handlerLogin(s *state, cmd command) error{
	if len(cmd.args) == 0{
		return fmt.Errorf("Error: Expected 1 arg, a username") 
	}
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil{
	os.Exit(1)
	}

	err = config.SetUser(cmd.args[0],*s.config)
	if err != nil{
	return fmt.Errorf("handlerLogin Error: %w\n", err)
	}
	fmt.Println("Username set successfully")
	return nil
}

func newCommands() *commands{
	return &commands{
		handlers: make(map[string]func(*state, command)error)} 
}

func (c *commands) register(name string, handler func(*state,command)error){
	c.handlers[name] = handler 
}

func (c *commands) run(s *state, cmd command) error{
	handler, exists := c.handlers[cmd.name]
	if !exists{
	return fmt.Errorf("command %s not found", cmd.name)
	}
	return handler(s, cmd)
}


	
	

	


