package store
import ("database/sql";"encoding/json";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type MockService struct{ID string `json:"id"`;Name string `json:"name"`;Prefix string `json:"prefix"`;Description string `json:"description,omitempty"`;CreatedAt string `json:"created_at"`;EndpointCount int `json:"endpoint_count"`;RequestCount int `json:"request_count"`}
type Endpoint struct{ID string `json:"id"`;ServiceID string `json:"service_id"`;Method string `json:"method"`;Path string `json:"path"`;StatusCode int `json:"status_code"`;ResponseBody string `json:"response_body"`;ResponseHeaders map[string]string `json:"response_headers,omitempty"`;DelayMs int `json:"delay_ms"`;Enabled bool `json:"enabled"`;CreatedAt string `json:"created_at"`}
type RequestLog struct{ID string `json:"id"`;EndpointID string `json:"endpoint_id"`;Method string `json:"method"`;Path string `json:"path"`;Headers string `json:"headers,omitempty"`;Body string `json:"body,omitempty"`;IP string `json:"ip,omitempty"`;StatusSent int `json:"status_sent"`;CreatedAt string `json:"created_at"`}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"mirage.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
for _,q:=range[]string{
`CREATE TABLE IF NOT EXISTS services(id TEXT PRIMARY KEY,name TEXT NOT NULL,prefix TEXT DEFAULT '',description TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`,
`CREATE TABLE IF NOT EXISTS endpoints(id TEXT PRIMARY KEY,service_id TEXT NOT NULL,method TEXT DEFAULT 'GET',path TEXT NOT NULL,status_code INTEGER DEFAULT 200,response_body TEXT DEFAULT '',response_headers TEXT DEFAULT '{}',delay_ms INTEGER DEFAULT 0,enabled INTEGER DEFAULT 1,created_at TEXT DEFAULT(datetime('now')))`,
`CREATE TABLE IF NOT EXISTS request_log(id TEXT PRIMARY KEY,endpoint_id TEXT DEFAULT '',method TEXT DEFAULT '',path TEXT DEFAULT '',headers TEXT DEFAULT '',body TEXT DEFAULT '',ip TEXT DEFAULT '',status_sent INTEGER DEFAULT 200,created_at TEXT DEFAULT(datetime('now')))`,
`CREATE INDEX IF NOT EXISTS idx_ep_service ON endpoints(service_id)`,
`CREATE INDEX IF NOT EXISTS idx_log_ep ON request_log(endpoint_id)`,
}{if _,err:=db.Exec(q);err!=nil{return nil,fmt.Errorf("migrate: %w",err)}};return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}

func(d *DB)CreateService(s *MockService)error{s.ID=genID();s.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO services VALUES(?,?,?,?,?)`,s.ID,s.Name,s.Prefix,s.Description,s.CreatedAt);return err}
func(d *DB)GetService(id string)*MockService{var s MockService;if d.db.QueryRow(`SELECT id,name,prefix,description,created_at FROM services WHERE id=?`,id).Scan(&s.ID,&s.Name,&s.Prefix,&s.Description,&s.CreatedAt)!=nil{return nil}
d.db.QueryRow(`SELECT COUNT(*) FROM endpoints WHERE service_id=?`,id).Scan(&s.EndpointCount)
d.db.QueryRow(`SELECT COUNT(*) FROM request_log WHERE endpoint_id IN(SELECT id FROM endpoints WHERE service_id=?)`,id).Scan(&s.RequestCount);return &s}
func(d *DB)ListServices()[]MockService{rows,_:=d.db.Query(`SELECT id,name,prefix,description,created_at FROM services ORDER BY name`);if rows==nil{return nil};defer rows.Close()
var o []MockService;for rows.Next(){var s MockService;rows.Scan(&s.ID,&s.Name,&s.Prefix,&s.Description,&s.CreatedAt)
d.db.QueryRow(`SELECT COUNT(*) FROM endpoints WHERE service_id=?`,s.ID).Scan(&s.EndpointCount);o=append(o,s)};return o}
func(d *DB)DeleteService(id string)error{d.db.Exec(`DELETE FROM request_log WHERE endpoint_id IN(SELECT id FROM endpoints WHERE service_id=?)`,id);d.db.Exec(`DELETE FROM endpoints WHERE service_id=?`,id);_,err:=d.db.Exec(`DELETE FROM services WHERE id=?`,id);return err}

func(d *DB)CreateEndpoint(e *Endpoint)error{e.ID=genID();e.CreatedAt=now();if e.Method==""{e.Method="GET"};if e.StatusCode==0{e.StatusCode=200};if e.ResponseHeaders==nil{e.ResponseHeaders=map[string]string{}}
hj,_:=json.Marshal(e.ResponseHeaders);en:=1;if !e.Enabled{en=0}
_,err:=d.db.Exec(`INSERT INTO endpoints VALUES(?,?,?,?,?,?,?,?,?,?)`,e.ID,e.ServiceID,e.Method,e.Path,e.StatusCode,e.ResponseBody,string(hj),e.DelayMs,en,e.CreatedAt);return err}
func(d *DB)GetEndpoint(id string)*Endpoint{var e Endpoint;var hj string;var en int
if d.db.QueryRow(`SELECT id,service_id,method,path,status_code,response_body,response_headers,delay_ms,enabled,created_at FROM endpoints WHERE id=?`,id).Scan(&e.ID,&e.ServiceID,&e.Method,&e.Path,&e.StatusCode,&e.ResponseBody,&hj,&e.DelayMs,&en,&e.CreatedAt)!=nil{return nil}
json.Unmarshal([]byte(hj),&e.ResponseHeaders);e.Enabled=en==1;return &e}
func(d *DB)ListEndpoints(serviceID string)[]Endpoint{rows,_:=d.db.Query(`SELECT id,service_id,method,path,status_code,response_body,response_headers,delay_ms,enabled,created_at FROM endpoints WHERE service_id=? ORDER BY method,path`,serviceID);if rows==nil{return nil};defer rows.Close()
var o []Endpoint;for rows.Next(){var e Endpoint;var hj string;var en int;rows.Scan(&e.ID,&e.ServiceID,&e.Method,&e.Path,&e.StatusCode,&e.ResponseBody,&hj,&e.DelayMs,&en,&e.CreatedAt)
json.Unmarshal([]byte(hj),&e.ResponseHeaders);e.Enabled=en==1;o=append(o,e)};return o}
func(d *DB)DeleteEndpoint(id string)error{d.db.Exec(`DELETE FROM request_log WHERE endpoint_id=?`,id);_,err:=d.db.Exec(`DELETE FROM endpoints WHERE id=?`,id);return err}
func(d *DB)MatchEndpoint(method, path string)*Endpoint{var e Endpoint;var hj string;var en int
if d.db.QueryRow(`SELECT id,service_id,method,path,status_code,response_body,response_headers,delay_ms,enabled,created_at FROM endpoints WHERE method=? AND path=? AND enabled=1`,method,path).Scan(&e.ID,&e.ServiceID,&e.Method,&e.Path,&e.StatusCode,&e.ResponseBody,&hj,&e.DelayMs,&en,&e.CreatedAt)!=nil{return nil}
json.Unmarshal([]byte(hj),&e.ResponseHeaders);e.Enabled=en==1;return &e}

func(d *DB)LogRequest(r *RequestLog)error{r.ID=genID();r.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO request_log VALUES(?,?,?,?,?,?,?,?,?)`,r.ID,r.EndpointID,r.Method,r.Path,r.Headers,r.Body,r.IP,r.StatusSent,r.CreatedAt);return err}
func(d *DB)ListRequests(endpointID string,limit int)[]RequestLog{if limit<=0{limit=50};rows,_:=d.db.Query(`SELECT id,endpoint_id,method,path,headers,body,ip,status_sent,created_at FROM request_log WHERE endpoint_id=? ORDER BY created_at DESC LIMIT ?`,endpointID,limit);if rows==nil{return nil};defer rows.Close()
var o []RequestLog;for rows.Next(){var r RequestLog;rows.Scan(&r.ID,&r.EndpointID,&r.Method,&r.Path,&r.Headers,&r.Body,&r.IP,&r.StatusSent,&r.CreatedAt);o=append(o,r)};return o}
func(d *DB)RecentRequests(limit int)[]RequestLog{if limit<=0{limit=50};rows,_:=d.db.Query(`SELECT id,endpoint_id,method,path,headers,body,ip,status_sent,created_at FROM request_log ORDER BY created_at DESC LIMIT ?`,limit);if rows==nil{return nil};defer rows.Close()
var o []RequestLog;for rows.Next(){var r RequestLog;rows.Scan(&r.ID,&r.EndpointID,&r.Method,&r.Path,&r.Headers,&r.Body,&r.IP,&r.StatusSent,&r.CreatedAt);o=append(o,r)};return o}

type Stats struct{Services int `json:"services"`;Endpoints int `json:"endpoints"`;Requests int `json:"requests"`}
func(d *DB)Stats()Stats{var s Stats;d.db.QueryRow(`SELECT COUNT(*) FROM services`).Scan(&s.Services);d.db.QueryRow(`SELECT COUNT(*) FROM endpoints`).Scan(&s.Endpoints);d.db.QueryRow(`SELECT COUNT(*) FROM request_log`).Scan(&s.Requests);return s}
