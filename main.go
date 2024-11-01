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
	"io"	
	"html"
	"net/http"
	"encoding/xml"
	"strconv"
	"strings"
	"os/exec"
	"runtime"	
   	"github.com/pressly/goose/v3"
)


type state struct{
	config *config.Config
	db *database.Queries
	rawDB *sql.DB
}

type command struct{
	name string
	args []string
}

type commands struct{
	handlers map[string]func(*state, command) error
}
 
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
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
	theState.rawDB = db
	defer db.Close()
	dbQueries := database.New(db)
	theState.db = dbQueries	
	theCommands := newCommands()
	theCommands.register("login", handlerLogin)
	theCommands.register("register", handlerRegister)
	theCommands.register("reset",handlerResetUsers)
	theCommands.register("users",middlewareLoggedIn(handlerUsers))	
	theCommands.register("addfeed",middlewareLoggedIn(addFeed))
	theCommands.register("feeds",handlerFeeds)
	theCommands.register("agg",agg)
	theCommands.register("follow",middlewareLoggedIn(follow))
	theCommands.register("following",middlewareLoggedIn(following))
	theCommands.register("unfollow",middlewareLoggedIn(unfollow))
	theCommands.register("browse",middlewareLoggedIn(handlerBrowse))
	theCommands.register("launch",middlewareLoggedIn(handlerLaunch))
	theCommands.register("migrate",migrate)
	theCommands.register("help",printHelp)
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
func migrate(s *state, cmd command)error{		
	migrationDir := "./sql/schema" 
	switch cmd.args[0]{
	case "up":
		err := goose.Up(s.rawDB,migrationDir)
		if err != nil{
			return err
		}
	case "down":
		err := goose.Down(s.rawDB,migrationDir)
		if err != nil{
			return err
		}

	default:
	fmt.Println("Choose a database migration either up or down")
	
	}
	return nil
}

func printHelp(_ *state, _ command)error{
	fmt.Println("Gator - RSS Feed Reader")
	fmt.Println("Commands:")
	fmt.Println("	login <username>		-logs you into the given user")
	fmt.Println("	register <username>		-registers you with the given username")
	fmt.Println("	agg [timeBetweenChecks]		-starts the aggregator witht he given time between checks")
	fmt.Println("	addfeed <FeedName> <FeedUrl>	-adds the feed at the url under your given name")
	fmt.Println("	browse [Amount]			-lists the posts(default 2)")
	fmt.Println("	launch <EntryNumber>		-opens the post in your default browser")
	fmt.Println("	reset				-reset registered users")
	fmt.Println("	feeds				-shows all feeds")
	fmt.Println("	follow <Url> 			-make this user follow the feed")
	fmt.Println("	following			-shows all feeds this user follows")
	fmt.Println("	unfollow <Url>			-unfollow a feed")
	fmt.Println("	users				-shows all users")
	fmt.Println("	migrate <up|down> 		-migrates the database up or down")
	fmt.Println("	help				-shows the commands")
	return nil
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
			return fmt.Errorf("username: %v already exists", newUserParams.Name)	
		}
	return err
	}
	err = config.SetUser(newUserParams.Name, *s.config)
	if err != nil{
	return err
	}
	fmt.Printf("%v successfully registered\n",newUser.Name)
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

func handlerUsers(s *state, _ command, user database.User) error{
	usersSlice,err := s.db.GetUsers(context.Background())	
	if err != nil{
	return err
	}
	currentUser := user
	for _ , user := range usersSlice{
		if user.Name == currentUser.Name{
			fmt.Printf("%v (current)\n",user.Name)
		}else{
			fmt.Printf("%v",user.Name)
		}
		
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string)(*RSSFeed, error){
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequestWithContext(ctx,"GET",feedURL,nil)
	if err != nil{
		return nil,err
	}
	req.Header.Add("User-Agent","gator")
	res, err := client.Do(req)
	if err != nil{
		return nil,err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil{
		return nil,err
	}
	responseFeed := RSSFeed{}
	err = xml.Unmarshal(data, &responseFeed)
	if err != nil{
		return nil,err
	}
	responseFeed.Channel.Title = html.UnescapeString(responseFeed.Channel.Title)
	responseFeed.Channel.Description = html.UnescapeString(responseFeed.Channel.Description)
	for index, item := range responseFeed.Channel.Item{
		responseFeed.Channel.Item[index].Title = html.UnescapeString(item.Title)
		responseFeed.Channel.Item[index].Description = html.UnescapeString(item.Description)
	}
	
	return &responseFeed, nil
}

func agg(s *state, cmd command)error{
	timeBetweenRequests,err:= time.ParseDuration("100s")
	if err != nil{
	return nil
	}
	if len(cmd.args) == 0 {
	 timeBetweenRequests,err = time.ParseDuration("30s")
		if err != nil{
		return err
		}
	}else{
	timeBetweenRequests,err = time.ParseDuration(cmd.args[0])
	if err != nil{
	return err
	}
	}
	fmt.Println(timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s,cmd)
		if err != nil{
		fmt.Println(err)
		}
	}
	
}
	
func addFeed(s *state, cmd command, user database.User)error{
	name := cmd.args[0]	
	if len(name) == 0{
		return fmt.Errorf("No blog name given")
	}
	url := cmd.args[1]
	if len(url) == 0{
		return fmt.Errorf("No URL given")
	}
	currentUser := user
	NewFeedParams := database.CreateFeedParams{
		ID:	uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: currentUser.ID,
		Name: name,
		Url: url,
		
	}


	newlyCreatedFeed, err := s.db.CreateFeed(context.Background(),NewFeedParams)
	if err != nil{
		return fmt.Errorf("addfeed Error: %w", err)
	}
		
	newFeedFollowParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: uuid.NullUUID{
			UUID: currentUser.ID,
			Valid: true,},
		FeedID: uuid.NullUUID{
			UUID: newlyCreatedFeed.ID,
			Valid: true,},
	}

	_,err = s.db.CreateFeedFollow(context.Background(),newFeedFollowParams)
		
	if err != nil{
	return fmt.Errorf("addFeed Follow error: %w\n", err)
	}
	return nil
}
	


func handlerFeeds(s *state, _ command) error{
	feedsSlice,err := s.db.GetFeeds(context.Background())	
	if err != nil{
	return err
	}	
	for _ , feed := range feedsSlice{
		userInfo, err := s.db.GetUserByID(context.Background(),feed.UserID)
		if err != nil{
			return err	
		}
		fmt.Printf("Name: %v\n",feed.Name)
		fmt.Printf("URL: %v\n",feed.Url)
		fmt.Printf("Created by: %v\n", userInfo.Name)
		fmt.Printf("Created: %v\n",feed.CreatedAt)
		fmt.Printf("Last Edited: %v\n", feed.UpdatedAt)
				
	}
	return nil
}

func follow(s *state, cmd command, user database.User)error{
	currentUserData := user
	feedToFollow, err := s.db.GetFeedByURL(context.Background(),cmd.args[0])
	if err != nil {
	return err
	}

	newFeedFollowParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: uuid.NullUUID{
			UUID: currentUserData.ID,
			Valid: true,},
		FeedID: uuid.NullUUID{
			UUID: feedToFollow.ID,
			Valid: true,},
	}

	feedFollowRow,err := s.db.CreateFeedFollow(context.Background(),newFeedFollowParams)
	if err != nil{
		return err
	}

	fmt.Printf("User: %v\n",feedFollowRow[0].UserName)
	fmt.Printf("Feed: %v\n",feedFollowRow[0].FeedName)
	return nil
	
}

func following(s *state, _ command, user database.User)error{
	followingData, err := s.db.GetFeedFollowsForUser(context.Background(),user.Name)
	if err != nil{
	return err
	}
	for _, row := range followingData{
	 fmt.Println(row.FeedName)
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User)error) func(s *state, cmd command)error{
	return func(s *state, cmd command)error{
	userLoggedIn, err := s.db.GetUser(context.Background(),s.config.CurrentUserName)
	if err != nil{
	return err
	}
		err = handler(s,cmd,userLoggedIn)
		if err != nil{
		return err
		}
	return nil
	}
	
}


func unfollow(s *state, cmd command, user database.User)error{

	newParams := database.DeleteFeedFollowByUserAndUrlParams{
		Name: user.Name,
		Url: cmd.args[0],

	}

	err := s.db.DeleteFeedFollowByUserAndUrl(context.Background(),newParams)
	if err != nil{
	return err
	}
	return nil
}
func getSqlNullTimeFromString(timestamp string)(sql.NullTime,error){
	timeFormats := []string{	
	 "Mon, 2 Jan 2006 15:04:05 GMT",
	"Mon, 02 Jan 2006 15:04:05 GMT",
	 "Mon, 2 Jan 2006 15:04:05 +0000",
	"Mon, 02 Jan 2008 15:04:05 +0000",}

var parsedTime time.Time
var err error


for _, format := range timeFormats{
	parsedTime, err = time.Parse(format, timestamp)
	if err == nil {
	break
	}
}
if err != nil{
	return sql.NullTime{},err
}
	return sql.NullTime{
		Time:parsedTime,
		Valid:true,
	},nil

}



func scrapeFeeds(s *state, _ command)error{	
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil{
	return err
	}
	MarkFeedParams := database.MarkFeedFetchedParams{
		LastFetchedAt: getCurrentTimeAsNullTime(), 
		ID: nextFeed.ID,
	}
	err = s.db.MarkFeedFetched(context.Background(),MarkFeedParams)
	if err != nil{
	return err
	}
	feedResponse, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
	return err
	}
	
	for _, item := range feedResponse.Channel.Item{	
		publicationTime,err := getSqlNullTimeFromString(item.PubDate)
		if err != nil{
		return err
		}
		postParams := database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title: getNullString(item.Title),
			Url: item.Link,
			Description: getNullString(item.Description),
			PublishedAt: publicationTime,
			FeedID:uuid.NullUUID{ 
				UUID: nextFeed.ID,
				Valid: true,
			},
		}
		
		_ ,err = s.db.CreatePost(context.Background(),postParams)	
		if err != nil{
			if strings.Contains(err.Error(),"duplicate key value violates unique constraint"){	
			continue	
			}
		return err
		}
		
	}

	return nil
}

func getNullString(inputString string) sql.NullString{
	return sql.NullString{
	String: inputString,
	Valid: true,
	}
}

func getCurrentTimeAsNullTime() sql.NullTime{
	return sql.NullTime{
		Time: time.Now(),
		Valid: true,
	}
}

func handlerBrowse(s *state, cmd command, user database.User)error{
	limitInt := 2
	if len(cmd.args) > 0{
		i,err := strconv.Atoi(cmd.args[0]) 
		if err != nil{
		return err
		}
		limitInt = i
	}
	NewGetPostParams := database.GetPostsForUserParams{
			UserID: uuid.NullUUID{
				UUID:user.ID,
				Valid: true,
		},
			Limit: int32(limitInt),

	}

	browseResponse,err := s.db.GetPostsForUser(context.Background(),NewGetPostParams)
	if err != nil{
	return err
	}
	for _,item := range browseResponse{
		fmt.Printf("Feed: %v\n",item.FeedName)
		fmt.Printf("Title: %v\n",item.Title.String)
		fmt.Printf("URL: %v\n",item.Url)
		if !strings.Contains(item.Description.String, item.Url){
			fmt.Printf("Description: %.100v\n",item.Description.String)
		}
		fmt.Printf("Published: %v\n", item.PublishedAt.Time)
		fmt.Printf("Post Number: %v\n", item.EntryNumber.Int32)
		fmt.Println("")
	}
	return nil

}

func handlerLaunch(s *state, cmd command, user database.User)error{
	if len(cmd.args) < 1{
	return fmt.Errorf("To load a post you need to enter the Post Number along with the launch command")
	}
	argConvertedFromString,err := strconv.Atoi(cmd.args[0])
	if err != nil{
	return err
	}
	convertedEntryNumber := sql.NullInt32{
			Int32: int32(argConvertedFromString),
			Valid: true,
	}
	postUrl, err := s.db.GetUrlByEntryNumber(context.Background(),convertedEntryNumber)
	err = openBrowser(postUrl)
	if err != nil{
		return err
	}
	return nil
}

func openBrowser(url string)error{
	var cmd string
	switch runtime.GOOS{
	case"darwin":
		cmd = "open"
	
	case "windows":
		cmd = "cmd"
	default:
		cmd = "xdg-open"
	}
	return exec.Command(cmd, url).Start()
}
