package draft

import (
	"sync"

	"github.com/gothing/draft/reflect"
)

// Scheme — описательная часть api и сбосов его использования
type Scheme struct {
	mu          sync.Mutex
	url         string
	name        string
	descr       string
	project     string
	cases       []*SchemeCase
	defAccess   AccessType
	defMethod   MethodType
	defConsumes MimeType
	defParams   interface{}
	defBody     interface{}
	defHeaders  SchemeCaseHeaders
	activeCase  *SchemeCase
}

// SchemeCase — описание и пример использование
type SchemeCase struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Access      AccessType        `json:"access"`
	Status      StatusType        `json:"status"`
	Method      MethodType        `json:"method"`
	Consumes    MimeType          `json:"consumes"`
	Params      interface{}       `json:"params"`
	Headers     SchemeCaseHeaders `json:"headers"`
	Body        interface{}       `json:"body"`
}

// SchemeCaseHeaders - описание хедеров
type SchemeCaseHeaders struct {
	Request  interface{} `json:"request"`
	Response interface{} `json:"response"`
}

// JSONScheme —
type JSONScheme struct {
	URL         string                           `json:"url"`
	Name        string                           `json:"name"`
	Project     string                           `json:"project"`
	Description string                           `json:"description"`
	Detail      map[StatusType]*JSONSchemeDetail `json:"detail"`
	Cases       []*SchemeCase                    `json:"cases"`
}

// JSONSchemeDetail —
type JSONSchemeDetail struct {
	Access   AccessType          `json:"access"`
	Request  *JSONSchemeRequest  `json:"request"`
	Response *JSONSchemeResponse `json:"response"`
}

// JSONSchemeRequest -
type JSONSchemeRequest struct {
	Method   MethodType              `json:"method"`
	Consumes MimeType                `json:"consumes"`
	Headers  map[string]reflect.Item `json:"headers"`
	Params   map[string]reflect.Item `json:"params"`
}

// JSONSchemeResponse -
type JSONSchemeResponse struct {
	Headers map[string]reflect.Item `json:"headers"`
	Body    map[string]reflect.Item `json:"body"`
}

// URL — относительный url
func (s *Scheme) URL(v string) {
	s.url = v
}

// Name — Нзвание конца
func (s *Scheme) Name(v string) {
	s.name = v
}

// Project — выставить права доступа к апишке или `case`
func (s *Scheme) Project(v string) {
	s.project = v
}

// Access — выставить права доступа к апишке или `case`
func (s *Scheme) Access(v AccessType) {
	if s.activeCase != nil {
		s.activeCase.Access = v
	} else {
		s.defAccess = v
	}
}

// Method — выставить метод к апишке или `case`
func (s *Scheme) Method(v MethodType) {
	if s.activeCase != nil {
		s.activeCase.Method = v
	} else {
		s.defMethod = v
	}
}

// Consumes — выставить Content-Type, принимаемый на вход, к апишке или `case`
func (s *Scheme) Consumes(v MimeType) {
	if s.activeCase != nil {
		s.activeCase.Consumes = v
	} else {
		s.defConsumes = v
	}
}

// Description — выставить описание к апишке или `case`
func (s *Scheme) Description(v string) {
	if s.activeCase != nil {
		s.activeCase.Description = v
	} else {
		s.descr = v
	}
}

// Params — выставить параметры к апишке или `case`
func (s *Scheme) Params(v interface{}) {
	if s.activeCase != nil {
		s.activeCase.Params = v
	} else {
		s.defParams = v
	}
}

// RequestHeaders — заголовки запроса
func (s *Scheme) RequestHeaders(v interface{}) {
	if s.activeCase != nil {
		s.activeCase.Headers.Request = v
	} else {
		s.defHeaders.Request = v
	}
}

// ResponseHeaders — заголовки ответа
func (s *Scheme) ResponseHeaders(v interface{}) {
	if s.activeCase != nil {
		s.activeCase.Headers.Response = v
	} else {
		s.defHeaders.Response = v
	}
}

// Body — выставить ответ к апишке или `case`
func (s *Scheme) Body(v interface{}) {
	if s.activeCase != nil {
		s.activeCase.Body = v
	} else {
		s.defBody = v
	}
}

// Case — определить описание и пример использования
func (s *Scheme) Case(status StatusType, name string, fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeCase = &SchemeCase{
		Status:   status,
		Name:     name,
		Method:   s.defMethod,
		Consumes: s.defConsumes,
		Access:   s.defAccess,
		Params:   s.defParams,
		Headers: SchemeCaseHeaders{
			Request:  s.defHeaders.Request,
			Response: s.defHeaders.Response,
		},
		Body: s.defBody,
	}

	s.cases = append(s.cases, s.activeCase)
	fn()
	s.activeCase = nil
}

// Cases —
func (s *Scheme) Cases() []*SchemeCase {
	return s.cases
}

// GetCaseByStatus —
func (s *Scheme) GetCaseByStatus(v StatusType) *SchemeCase {
	for _, c := range s.cases {
		if c.Status == v {
			return c
		}
	}
	return nil
}

// ToJSON — определить описание и пример использования
func (s *Scheme) ToJSON() JSONScheme {
	json := JSONScheme{
		URL:         s.url,
		Name:        s.name,
		Project:     s.project,
		Description: s.descr,
		Detail:      make(map[StatusType]*JSONSchemeDetail),
		Cases:       make([]*SchemeCase, len(s.cases)),
	}

	for i, c := range s.cases {
		d, exists := json.Detail[c.Status]
		if !exists {
			d = &JSONSchemeDetail{
				Request: &JSONSchemeRequest{
					Method:   c.Method,
					Consumes: c.Consumes,
					Headers:  make(map[string]reflect.Item),
					Params:   make(map[string]reflect.Item),
				},

				Response: &JSONSchemeResponse{
					Headers: make(map[string]reflect.Item),
					Body:    make(map[string]reflect.Item),
				},
			}

			json.Detail[c.Status] = d
		}

		d.Access = c.Access

		nc := &SchemeCase{}
		*nc = *c

		nc.Headers.Request = setReflectObjectMap(d.Request.Headers, c.Headers.Request)
		nc.Headers.Response = setReflectObjectMap(d.Response.Headers, c.Headers.Response)

		nc.Params = setReflectObjectMap(d.Request.Params, c.Params)
		nc.Body = setReflectObjectMap(d.Response.Body, c.Body)

		json.Cases[i] = nc
	}

	return json
}

func setReflectObjectMap(obj map[string]reflect.Item, v interface{}) interface{} {
	if v != nil {
		ref := reflect.Get(v, reflect.Options{
			SnakeCase: true,
		})
		for _, item := range ref.Nested {
			obj[item.Name] = item
		}

		return prepareMock(ref)
	}

	return nil
}
