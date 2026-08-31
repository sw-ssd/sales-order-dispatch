package server

// InitDomains 逐 domain 組裝 repo→usecase→handler 並掛上 router。
// 新增 domain 只動此檔（D31）；各 domain 於對應計畫落地。
func (s *Server) InitDomains() {
	s.mountAuth()
}
