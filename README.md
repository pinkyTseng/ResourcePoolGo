# ResourcePoolGo

## 主要思路
我是8/9有寄信請教題目的泛型functions,可能在Release時沒辦法判斷是否重覆對同一資源Release問題的同學,因為團隊回覆我可以直接做修改沒有問題,先附上我稍微修改後的定義

```
type Pool[T any] interface {  
	// This creates or returns a ready-to-use item from the resource pool  
	Acquire(context.Context) (*PoolResource[T], error)  
	// This releases an active resource back to the resource pool  
	Release(*PoolResource[T])  
	// This returns the number of idle items  
	NumIdle() int  
}  

func New[T any](  
	creator func(context.Context) (*PoolResource[T], error),  
	maxIdleSize int,  
	maxIdleTime time.Duration,  
) Pool[T]  

//新增的
type PoolResource[T any] struct {
	value T
	timer *time.Timer
	id    uintptr
}
//實際Pool struct,題目外的field公用請參考註解 
type GenericPool[T any] struct {	
	creator       func(context.Context) (*PoolResource[T], error)
	maxIdleSize   int
	maxIdleTime   time.Duration
    
    //將idle resources存入map
    idleObjects   *collections.SyncIdentityMap
    //將active(正被取用中) resources存入map,release時才能避免重複問題另外也可計算和限制pool MaxSize
	activeObjects *collections.SyncIdentityMap
	//主要是當idleObjects的queue用 會從較舊的資源先取 
    PoolIdArr     *collections.SyncArr
	//idleObjects size + activeObjects size,主要是讓pool總量能受到限制,不然很容易會有資源爆炸的問題
    MaxSize       int
    //防止race condition用
	globalMtx     *sync.RWMutex
}
```

在一開始我本來是想只記錄idleObjects pool,當Release被呼叫時只要maxIdleSize未滿就將resource obj放入idleObjects pool,透過這種方式來達到最開始題目給的任意泛型的目的,但後來我想到這樣做的話假使同一個資源一直重覆呼叫Release,這做法就會在未達到maxIdleSize時一直往idleObjectsl放入resource obj,在稍做修改functions泛型定義後,因為PoolResource struct 包含id,如此就可避免重複放入idle pool!另外,id的部分我本來打算用UUID,但因為pdf有提到盡量不要使用非內建package,所以我改用任意物件都有的uintptr!  

- 這邊的Generic Resource Pool的功能我主要設計如下
    - 當Acquire被呼叫,idleObjects有resource的話直接取出最舊的返回給caller,當idleObjects沒有resource時產生新的resource返回給caller   
    - 當Acquire被呼叫,idleObjects有resource的話直接取出最舊的返回給caller,並參考Slice動態擴容的概念,會另開一個Goroutine只要不超過MaxSize和maxIdleSize會自動呼叫creator創建一個resource object放到idleObjects中供下次快速取用
    - 當Release被呼叫,若activeObjects不含release的資源則代表重複release直接返回不做處理
    - 當Release被呼叫,resource object從activeObjects移出,只要不超過maxIdleSize放到idleObjects中,若超過則丟棄不處理
    - 當resource object被放入idleObjects中,timer就會開始計算,當達maxIdleTime會將其從idleObjects中移除

## 本次未納入的一些想法
- 創建資源超時取消功能,因為題目的Acquire和creator function都有Context參數,Context的一個重要功能就在能當超時時,進行一些中間資源的清理和一些處理,但是因為題目沒提到有timeout相關設定,我覺得可能相對來說不是這題目的重點專注處,避免過度設計就沒納入
- 因為pool常用在connection的建立,本來有思考到實際接mysql之類的設計,但因題目有提到盡量不要使用非內建package且時間有限也就沒將其納入
- 因為pool常用在有url的connection的建立和一些建立相對耗時或耗資源的物件上,有思考到將其抽象成有url和耗資源的２大類,並在其外再封裝一層以便直接使用這2大類的resource pool,但題目是要求generic resource pool且creator是讓調用者自行傳入,且原始題目是可傳入任意型別的泛型,我覺得對各別實際類型的分類封裝可能相對來說不是這題目的重點專注處且時間有限也就沒將其納入
- 一開始本來只打算記idleObjects沒打算記activeObjects時idleObjects本來是用channel保存,但後來考慮到activeObjects要Release時需要知道被Release的是哪個資源所以採用map,當時直覺想說統一資料型態比較方便所以idleObjects也改用map保存,但是後來思考了一下想到idleObjects可能還是用channel保存會更高效,因為channel本身特性,這樣就不用保存PoolIdArr和進行其操作

## 執行步驟
這work基本上是呼叫New function即可操作使用該pool,主要是執行go test來驗證pool的使用效果和正確性,因為各別function主要是new或取出或放回資源,且要配合Resource pool現在的狀態較有意義,我沒針對單獨單一function寫test case,我主要是針對一些實踐很可能會用到的使用情境及邊界狀況寫test case,另外,test case中我有考慮實際使用pool的情況絕大多數會是在多線程(協程)的環境下使用,所以我有些test case也有用Goroutine來做測試,function名以MG結尾的Test function就是有考慮並發的test case,其他的從function名應該就看得出來主要test case是要測什麼情境,另外,測試主要使用的2種resource型態是string和ConnectResource struct,但如上面所說,這邊的resource並沒有真的有實作DB connection的功能和較耗資源的創建操作等