package api

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"appengine/user"

	"github.com/crhym3/go-endpoints/endpoints"
	
	//"appengine/datastore"

	//"config"

)


//const clientId = "YOUR-CLIENT-ID"
//var clientId = config.Config.OAuthProviders.Google.ClientId
var clientId = "882975820932-q34i2m1lklcmv8kqqrcleumtdhe4qbhk.apps.googleusercontent.com"


var (
	scopes    = []string{endpoints.EmailScope}
	clientIds = []string{clientId, endpoints.ApiExplorerClientId}
	// in case we'll want to use Mindale API from an Android app
	audiences = []string{clientId}
)

type BoardMsg struct {
	State string "json:'state' endpoints:'required'"
}

type ScoreReqMsg struct {
	Outcome string "json:'outcome' endpoints:'required'"
}

type ScoreRespMsg struct {
	Id      int64  `json:"id"`
	Outcome string `json:"outcome"`
	Played  string `json:"played"`
}

type ScoresListReq struct {
	Limit int "json:'limit'"
}

type ScoresListResp struct {
	Items []*ScoreRespMsg "json:'items'"
}

// Mindale API service
type ServiceApi struct {
}

// BoardGetMove simulates a computer move in mindale.
// Exposed as API endpoint
func (ttt *ServiceApi) BoardGetMove(r *http.Request,
	req *BoardMsg, resp *BoardMsg) error {

	const boardLen = 9
	if len(req.State) != boardLen {
		return fmt.Errorf("Bad Request: Invalid board: %q", req.State)
	}
	runes := []rune(req.State)
	freeIndices := make([]int, 0)
	for pos, r := range runes {
		if r != 'O' && r != 'X' && r != '-' {
			return fmt.Errorf("Bad Request: Invalid rune: %q", r)
		}
		if r == '-' {
			freeIndices = append(freeIndices, pos)
		}
	}
	freeIdxLen := len(freeIndices)
	if freeIdxLen > 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomIdx := r.Intn(freeIdxLen)
		runes[freeIndices[randomIdx]] = 'O'
		resp.State = string(runes)
	} else {
		return fmt.Errorf("Bad Request: This board is full: %q", req.State)
	}
	return nil
}

// ScoresList queries scores for the current user.
// Exposed as API endpoint
func (ttt *ServiceApi) ScoresList(r *http.Request,
	req *ScoresListReq, resp *ScoresListResp) error {

	c := endpoints.NewContext(r)
	u, err := getCurrentUser(c)
	if err != nil {
		return err
	}
	q := newUserScoreQuery(u)
	if req.Limit <= 0 {
		req.Limit = 10
	}
	scores, err := fetchScores(c, q, req.Limit)
	if err != nil {
		return err
	}
	resp.Items = make([]*ScoreRespMsg, len(scores))
	for i, score := range scores {
		resp.Items[i] = score.toMessage(nil)
	}
	return nil
}

// ScoresInsert inserts a new score for the current user.
func (ttt *ServiceApi) ScoresInsert(r *http.Request,
	req *ScoreReqMsg, resp *ScoreRespMsg) error {

	c := endpoints.NewContext(r)
	u, err := getCurrentUser(c)
	if err != nil {
		return err
	}
	score := newScore(req.Outcome, u)
	if err := score.put(c); err != nil {
		return err
	}
	score.toMessage(resp)
	return nil
}

// getCurrentUser retrieves a user associated with the request.
// If there's no user (e.g. no auth info present in the request) returns
// an "unauthorized" error.
func getCurrentUser(c endpoints.Context) (*user.User, error) {

	user, err := endpoints.CurrentUser(c, scopes, audiences, clientIds)
	if err != nil {
		c.Errorf("User not found: %s", err)
		return nil, err
	}

	if user == nil {
		return nil, errors.New("Unauthorized: Please, sign in.")
	}

	c.Debugf("Current user: %s", user)
	return user, nil
}




func createMethod(rpcService *endpoints.RpcService, service, path, method, name string){

	info := rpcService.MethodByName(name).Info()
	info.Path, info.HttpMethod, info.Name = path, method, name
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences
}



// RegisterService exposes ServiceApi methods as API endpoints.
// 
// The registration/initialization during startup is not performed here but
// in app package. It is separated from this package (mindale) so that the
// service and its methods defined here can be used in another app,
// e.g. http://github.com/crhym3/go-endpoints.appspot.com.
func RegisterService() (*endpoints.RpcService, error) {
	api := &ServiceApi{}
	rpcService, err := endpoints.RegisterService(api,
		"mindale", "v1", "Mindale API", true)
	if err != nil {
		return nil, err
	}



	info := rpcService.MethodByName("PaymentsAdd").Info()
	info.Path, info.HttpMethod, info.Name = "payments", "GET", "payments.add"
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences


	info = rpcService.MethodByName("PaymentsList").Info()
	info.Path, info.HttpMethod, info.Name = "payments", "GET", "payments.list"
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences





	info = rpcService.MethodByName("BoardGetMove").Info()
	info.Path, info.HttpMethod, info.Name = "board", "POST", "board.getmove"
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences

	info = rpcService.MethodByName("ScoresList").Info()
	info.Path, info.HttpMethod, info.Name = "scores", "GET", "scores.list"
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences

	info = rpcService.MethodByName("ScoresInsert").Info()
	info.Path, info.HttpMethod, info.Name = "scores", "POST", "scores.insert"
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences


	return rpcService, nil
}