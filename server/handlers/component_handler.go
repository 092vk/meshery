package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/layer5io/meshery/server/helpers"
	"github.com/layer5io/meshery/server/helpers/utils"
	"github.com/layer5io/meshery/server/models"
	"github.com/layer5io/meshery/server/models/pattern/core"
	"github.com/layer5io/meshkit/models/events"
	"github.com/layer5io/meshkit/models/oci"
	"github.com/layer5io/meshkit/models/registration"
	meshkitutils "github.com/layer5io/meshkit/utils"

	_models "github.com/layer5io/meshkit/models/meshmodel/core/v1beta1"
	"github.com/meshery/schemas/models/v1alpha3/relationship"
	"github.com/meshery/schemas/models/v1beta1/component"
	"github.com/meshery/schemas/models/v1beta1/connection"
	"github.com/meshery/schemas/models/v1beta1/model"
	_model "github.com/meshery/schemas/models/v1beta1/model"

	"github.com/layer5io/meshkit/models/meshmodel/entity"
	"github.com/layer5io/meshkit/models/meshmodel/registry"

	regv1beta1 "github.com/layer5io/meshkit/models/meshmodel/registry/v1beta1"
)

/**Meshmodel endpoints **/
const DefaultPageSizeForMeshModelComponents = 25

// swagger:route GET /api/meshmodels/categories/{category}/models GetMeshmodelModelsByCategories idGetMeshmodelModelsByCategories
//
// Handle GET request for getting all meshmodel models for a given category. The component type/model name should be lowercase like "kubernetes", "istio"
//
// ```?version={version}``` If version is unspecified then all models are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?search={modelname}``` If search is non empty then a greedy search is performed
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
// ```?annotations={["true"/"false"/]}``` When this query parameter is "true", only models with the "isAnnotation" property set to true are returned. When  this query parameter is "false", all models except those considered to be annotation models are returned. Any other value of the query parameter results in both annoations as well as non-annotation models being returned.
// responses:
// ```?annotations={["true"/"false"/]}``` If "true" models having "isAnnotation" property as true are "only" returned, If false all models except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation models being returned.
//
//	200: []meshmodelModelsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelModelsByCategories(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	cat := mux.Vars(r)["category"]
	queryParams := r.URL.Query()
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	returnAnnotationComp := queryParams.Get("annotations")

	filter := &regv1beta1.ModelFilter{
		Category:    cat,
		Version:     queryParams.Get("version"),
		Limit:       limit,
		Offset:      offset,
		OrderOn:     order,
		Sort:        sort,
		Annotations: returnAnnotationComp,
	}
	if search != "" {
		filter.Greedy = true
		filter.DisplayName = search
	}
	entities, count, _, _ := h.registryManager.GetEntities(filter)
	var modelDefs []_model.ModelDefinition
	for _, model := range entities {
		model, ok := model.(*_model.ModelDefinition)
		if ok {
			modelDefs = append(modelDefs, *model)
		}
	}

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	res := models.MeshmodelsDuplicateAPIResponse{
		Page:     page,
		PageSize: int(pgSize),
		Count:    count,
		Models:   models.FindDuplicateModels(modelDefs),
	}

	if err := enc.Encode(res); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/categories/{category}/models/{model} GetMeshmodelModelsByCategoriesByModel idGetMeshmodelModelsByCategoriesByModel
//
// Handle GET request for getting all meshmodel models for a given category. The component type/model name should be lowercase like "kubernetes", "istio"
//
// ```?version={version}``` If version is unspecified then all models are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?search={[true/false]}``` If search is true then a greedy search is performed
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
// responses:
// ```?annotations={["true"/"false"/]}``` If "true" models having "isAnnotation" property as true are "only" returned, If false all models except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation models being returned.
//
//	200: []meshmodelModelsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelModelsByCategoriesByModel(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	cat := mux.Vars(r)["category"]
	model := mux.Vars(r)["model"]
	queryParams := r.URL.Query()
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	var greedy bool
	if search == "true" {
		greedy = true
	}
	returnAnnotationComp := queryParams.Get("annotations")

	entities, count, _, _ := h.registryManager.GetEntities(&regv1beta1.ModelFilter{
		Category:    cat,
		Name:        model,
		Version:     queryParams.Get("version"),
		Limit:       limit,
		Offset:      offset,
		Greedy:      greedy,
		OrderOn:     order,
		Sort:        sort,
		Annotations: returnAnnotationComp,
	})

	var modelDefs []_model.ModelDefinition
	for _, model := range entities {
		model, ok := model.(*_model.ModelDefinition)
		if ok {
			modelDefs = append(modelDefs, *model)
		}
	}

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	res := models.MeshmodelsDuplicateAPIResponse{
		Page:     page,
		PageSize: int(pgSize),
		Count:    count,
		Models:   models.FindDuplicateModels(modelDefs),
	}

	if err := enc.Encode(res); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/models GetMeshmodelModels idGetMeshmodelModels
// Handle GET request for getting all meshmodel models
//
// # Returns a list of registered models across all categories
//
// ```?version={version}``` If version is unspecified then all models are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?search={modelname}``` If search is non empty then a greedy search is performed
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
// responses:
// ```?annotations={["true"/"false"/]}``` If "true" models having "isAnnotation" property as true are "only" returned, If false all models except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation models being returned.
//
//	200: meshmodelModelsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelModels(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	queryParams := r.URL.Query()
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	v := queryParams.Get("version")
	returnAnnotationComp := queryParams.Get("annotations")

	filter := &regv1beta1.ModelFilter{
		Registrant:  queryParams.Get("registrant"),
		Version:     v,
		Limit:       limit,
		Offset:      offset,
		OrderOn:     order,
		Sort:        sort,
		Annotations: returnAnnotationComp,

		Components:    queryParams.Get("components") == "true",
		Relationships: queryParams.Get("relationships") == "true",
		Status:        queryParams.Get("status"),
	}
	if search != "" {
		filter.DisplayName = search
		filter.Greedy = true
	}

	entities, count, _, _ := h.registryManager.GetEntities(filter)
	var modelDefs []_model.ModelDefinition
	for _, model := range entities {
		model, ok := model.(*_model.ModelDefinition)
		if ok {
			modelDefs = append(modelDefs, *model)
		}
	}
	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	res := models.MeshmodelsDuplicateAPIResponse{
		Page:     page,
		PageSize: int(pgSize),
		Count:    count,
		Models:   models.FindDuplicateModels(modelDefs),
	}

	if err := enc.Encode(res); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/models/{model} GetMeshmodelModelsByName idGetMeshmodelModelsByName
// Handle GET request for getting all meshmodel models. The component type/model name should be lowercase like "kubernetes", "istio"
//
// # Returns a list of registered models across all categories
//
// ```?version={version}``` If version is unspecified then all models are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?search={[true/false]}``` If search is true then a greedy search is performed
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
// responses:
// ```?annotations={["true"/"false"/]}``` If "true" models having "isAnnotation" property as true are "only" returned, If false all models except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation models being returned.
//
//	200: []meshmodelModelsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelModelsByName(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	name := mux.Vars(r)["model"]
	queryParams := r.URL.Query()
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	var greedy bool
	if search == "true" {
		greedy = true
	}
	v := queryParams.Get("version")
	returnAnnotationComp := queryParams.Get("annotations")
	entities, count, _, _ := h.registryManager.GetEntities(&regv1beta1.ModelFilter{
		Name:        name,
		Version:     v,
		Limit:       limit,
		Offset:      offset,
		Greedy:      greedy,
		OrderOn:     order,
		Sort:        sort,
		Annotations: returnAnnotationComp,

		Components:    queryParams.Get("components") == "true",
		Relationships: queryParams.Get("relationships") == "true",
	})

	var modelDefs []_model.ModelDefinition
	for _, model := range entities {
		model, ok := model.(*_model.ModelDefinition)
		if ok {
			modelDefs = append(modelDefs, *model)
		}
	}

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	res := models.MeshmodelsDuplicateAPIResponse{
		Page:     page,
		PageSize: int(pgSize),
		Count:    count,
		Models:   models.FindDuplicateModels(modelDefs),
	}

	if err := enc.Encode(res); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/categories GetMeshmodelCategories idGetMeshmodelCategories
// Handle GET request for getting all meshmodel categories
//
// ```?order={field}``` orders on the passed field
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?search={categoryName}``` If search is non empty then a greedy search is performed
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
// responses:
//
//	200: []meshmodelCategoriesResponseWrapper
func (h *Handler) GetMeshmodelCategories(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	filter := &regv1beta1.CategoryFilter{
		Limit:   limit,
		Offset:  offset,
		OrderOn: order,
		Sort:    sort,
	}
	if search != "" {
		filter.Greedy = true
		filter.Name = search
	}

	categories, count, _, _ := h.registryManager.GetEntities(filter)

	var pgSize int64

	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	res := models.MeshmodelCategoriesAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Categories: categories,
	}

	if err := enc.Encode(res); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/categories/{category} GetMeshmodelCategoriesByName idGetMeshmodelCategoriesByName
// Handle GET request for getting all meshmodel categories of a given name
//
// ```?order={field}``` orders on the passed field
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// ```?search={[true/false]}``` If search is true then a greedy search is performed
// responses:
//
//	200: []meshmodelCategoriesResponseWrapper
func (h *Handler) GetMeshmodelCategoriesByName(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	name := mux.Vars(r)["category"]
	var greedy bool
	if search == "true" {
		greedy = true
	}
	categories, count, _, _ := h.registryManager.GetEntities(&regv1beta1.CategoryFilter{
		Name:    name,
		Limit:   limit,
		Greedy:  greedy,
		Offset:  offset,
		OrderOn: order,
		Sort:    sort,
	})

	var pgSize int64

	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	res := models.MeshmodelCategoriesAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Categories: categories,
	}

	if err := enc.Encode(res); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/categories/{category}/models/{model}/components/{name} GetMeshmodelComponentsByNameByModelByCategory idGetMeshmodelComponentsByNameByModelByCategory
// Handle GET request for getting meshmodel components of a specific type by model and category.
//
// Example: ```/api/meshmodels/categories/Orchestration``` and Management/models/kubernetes/components/Namespace
// Components can be further filtered through query parameter
//
// ```?version={version}``` If version is unspecified then all model versions are returned
//
// ```?apiVersion={apiVersion}``` If apiVersion is unspecified then all components are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?search={[true/false]}``` If search is true then a greedy search is performed
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// ```?annotations={["true"/"false"/]}``` If "true" components having "isAnnotation" property as true are "only" returned, If false all components except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation components being returned.
// responses:
// 200: []meshmodelComponentsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelComponentsByNameByModelByCategory(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	name := mux.Vars(r)["name"]

	queryParams := r.URL.Query()
	var greedy bool
	if search == "true" {
		greedy = true
	}
	typ := mux.Vars(r)["model"]
	cat := mux.Vars(r)["category"]
	v := queryParams.Get("version")
	returnAnnotationComp := queryParams.Get("annotations")
	entities, count, _, _ := h.registryManager.GetEntities(&regv1beta1.ComponentFilter{
		Name:         name,
		CategoryName: cat,
		ModelName:    typ,
		APIVersion:   queryParams.Get("apiVersion"),
		Version:      v,
		Offset:       offset,
		Greedy:       greedy,
		Limit:        limit,
		OrderOn:      order,
		Sort:         sort,
		Annotations:  returnAnnotationComp,
	})

	comps := prettifyCompDefSchema(entities)

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	response := models.MeshmodelComponentsDuplicateAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Components: models.FindDuplicateComponents(comps),
	}

	if err := enc.Encode(response); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/categories/{category}/components/{name} GetMeshmodelComponentsByNameByCategory idGetMeshmodelComponentsByNameByCategory
// Handle GET request for getting meshmodel components of a specific type category.
//
// Example: ```/api/meshmodels/categories/Orchestration``` and Management/components/Namespace
// Components can be further filtered through query parameter
//
// ```?model={model}``` If model is unspecified then all models are returned
//
// ```?version={version}``` If version is unspecified then all model versions are returned
//
// ```?apiVersion={apiVersion}``` If apiVersion is unspecified then all components are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?search={[true/false]}``` If search is true then a greedy search is performed
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// ```?annotations={["true"/"false"/]}``` If "true" components having "isAnnotation" property as true are "only" returned, If false all components except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation components being returned.
// responses:
//
//	200: []meshmodelComponentsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelComponentsByNameByCategory(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	name := mux.Vars(r)["name"]
	var greedy bool
	queryParams := r.URL.Query()
	if search == "true" {
		greedy = true
	}
	cat := mux.Vars(r)["category"]
	v := queryParams.Get("version")
	returnAnnotationComp := queryParams.Get("annotations")

	entities, count, _, _ := h.registryManager.GetEntities(&regv1beta1.ComponentFilter{
		Name:         name,
		ModelName:    queryParams.Get("model"),
		CategoryName: cat,
		APIVersion:   queryParams.Get("apiVersion"),
		Version:      v,
		Offset:       offset,
		Limit:        limit,
		Greedy:       greedy,
		OrderOn:      order,
		Sort:         sort,
		Annotations:  returnAnnotationComp,
	})
	comps := prettifyCompDefSchema(entities)

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	response := models.MeshmodelComponentsDuplicateAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Components: models.FindDuplicateComponents(comps),
	}

	if err := enc.Encode(response); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/models/{model}/components/{name} GetMeshmodelComponentsByNameByModel idGetMeshmodelComponentsByNameByModel
// Handle GET request for getting meshmodel components of a specific  model.
//
// Example: ```/api/meshmodels/models/kubernetes/components/Namespace```
// Components can be further filtered through query parameter
//
// ```?version={version}``` If version is unspecified then all model versions are returned
//
// ```?apiVersion={apiVersion}``` If apiVersion is unspecified then all components are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?search={[true/false]}``` If search is true then a greedy search is performed
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// ```?annotations={["true"/"false"/]}``` If "true" components having "isAnnotation" property as true are "only" returned, If false all components except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation components being returned. returned.
// responses:
//
//	200: []meshmodelComponentsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelComponentsByNameByModel(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	name := mux.Vars(r)["name"]
	var greedy bool
	queryParams := r.URL.Query()

	if search == "true" {
		greedy = true
	}
	typ := mux.Vars(r)["model"]
	v := queryParams.Get("version")

	returnAnnotationComp := queryParams.Get("annotations")

	entities, count, _, _ := h.registryManager.GetEntities(&regv1beta1.ComponentFilter{
		Name:        name,
		ModelName:   typ,
		APIVersion:  queryParams.Get("apiVersion"),
		Version:     v,
		Offset:      offset,
		Greedy:      greedy,
		Limit:       limit,
		OrderOn:     order,
		Sort:        sort,
		Annotations: returnAnnotationComp,
	})
	comps := prettifyCompDefSchema(entities)

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	response := models.MeshmodelComponentsDuplicateAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Components: models.FindDuplicateComponents(comps),
	}

	if err := enc.Encode(response); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/components/{name} GetAllMeshmodelComponentsByName idGetAllMeshmodelComponentsByName
// Handle GET request for getting meshmodel components of a specific type by name across all models and categories
//
// Example: ```/api/meshmodels/components/Namespace```
// Components can be further filtered through query parameter
//
// ```?model={model}``` If model is unspecified then all models are returned
//
// ```?version={version}``` If version is unspecified then all model versions are returned
//
// ```?apiVersion={apiVersion}``` If apiVersion is unspecified then all components are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?trim={[true]}``` When trim is set to true, the underlying schemas are not returned for entities
//
// ```?search={[true/false]}``` If search is true then a greedy search is performed
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// ```?annotations={["true"/"false"/]}``` If "true" components having "isAnnotation" property as true are "only" returned, If false all components except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation components being returned.
// responses:
// 200: []meshmodelComponentsDuplicateResponseWrapper
func (h *Handler) GetAllMeshmodelComponentsByName(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	name := mux.Vars(r)["name"]
	var greedy bool
	queryParams := r.URL.Query()
	if search == "true" {
		greedy = true
	}
	v := queryParams.Get("version")
	returnAnnotationComp := queryParams.Get("annotations")
	entities, count, _, _ := h.registryManager.GetEntities(&regv1beta1.ComponentFilter{
		Name:        name,
		Trim:        queryParams.Get("trim") == "true",
		APIVersion:  queryParams.Get("apiVersion"),
		Version:     v,
		ModelName:   queryParams.Get("model"),
		Offset:      offset,
		Limit:       limit,
		Greedy:      greedy,
		OrderOn:     order,
		Sort:        sort,
		Annotations: returnAnnotationComp,
	})

	comps := prettifyCompDefSchema(entities)

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	response := models.MeshmodelComponentsDuplicateAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Components: models.FindDuplicateComponents(comps),
	}

	if err := enc.Encode(response); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/models/{model}/components GetMeshmodelComponentByModel idGetMeshmodelComponentByModel
// Handle GET request for getting meshmodel components of a specific model. The component type/model name should be lowercase like "kubernetes", "istio"
//
// Example: ```/api/meshmodels/models/kubernetes/components```
// Components can be further filtered through query parameter
//
// ```?version={version}```
//
// ```?trim={[true]}``` When trim is set to true, the underlying schemas are not returned for entities
//
// ```?apiVersion={apiVersion}``` If apiVersion is unspecified then all models are returned
//
// ```?search={componentname}``` If search is non empty then a greedy search is performed
//
// ```?order={field}``` orders on the passed field
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// ```?annotations={["true"/"false"/]}``` If "true" components having "isAnnotation" property as true are "only" returned, If false all components except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation components being returned.
// responses:
// 200: []meshmodelComponentsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelComponentByModel(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	typ := mux.Vars(r)["model"]
	queryParams := r.URL.Query()
	v := queryParams.Get("version")

	returnAnnotationComp := queryParams.Get("annotations")
	filter := &regv1beta1.ComponentFilter{
		ModelName:   typ,
		Version:     v,
		Trim:        queryParams.Get("trim") == "true",
		APIVersion:  queryParams.Get("apiVersion"),
		Limit:       limit,
		Offset:      offset,
		OrderOn:     order,
		Sort:        sort,
		Annotations: returnAnnotationComp,
	}
	if search != "" {
		filter.Greedy = true
		filter.DisplayName = search
	}
	entities, count, _, _ := h.registryManager.GetEntities(filter)
	comps := prettifyCompDefSchema(entities)

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	response := models.MeshmodelComponentsDuplicateAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Components: models.FindDuplicateComponents(comps),
	}

	if err := enc.Encode(response); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/categories/{category}/models/{model}/components GetMeshmodelComponentByModelByCategory idGetMeshmodelComponentByModelByCategory
//
// Handle GET request for getting meshmodel components of a specific model and category. The component type/model name should be lowercase like "kubernetes", "istio"
//
// Example: ```/api/meshmodels/categories/Orchestration``` and Management/models/kubernetes/components
// Components can be further filtered through query parameter
//
// ```?version={version}```
//
// ```?trim={[true]}``` When trim is set to true, the underlying schemas are not returned for entities
//
// ```?apiVersion={apiVersion}``` If apiVersion is unspecified then all models are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?search={componentname}``` If search is non empty then a greedy search is performed
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// ```?annotations={["true"/"false"/]}``` If "true" components having "isAnnotation" property as true are "only" returned, If false all components except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation components being returned.
// responses:
// 200: []meshmodelComponentsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelComponentByModelByCategory(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	typ := mux.Vars(r)["model"]
	cat := mux.Vars(r)["category"]
	queryParams := r.URL.Query()
	v := queryParams.Get("version")
	returnAnnotationComp := queryParams.Get("annotations")
	filter := &regv1beta1.ComponentFilter{
		CategoryName: cat,
		ModelName:    typ,
		Version:      v,
		Trim:         queryParams.Get("trim") == "true",
		APIVersion:   queryParams.Get("apiVersion"),
		Limit:        limit,
		Offset:       offset,
		OrderOn:      order,
		Sort:         sort,
		Annotations:  returnAnnotationComp,
	}
	if search != "" {
		filter.Greedy = true
		filter.DisplayName = search
	}
	entities, count, _, _ := h.registryManager.GetEntities(filter)
	comps := prettifyCompDefSchema(entities)

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	response := models.MeshmodelComponentsDuplicateAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Components: models.FindDuplicateComponents(comps),
	}

	if err := enc.Encode(response); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/categories/{category}/components GetMeshmodelComponentByCategory idGetMeshmodelComponentByCategory
// Handle GET request for getting meshmodel components of a specific model and category.
//
// # Components can be further filtered through query parameter
//
// ```?version={version}```
//
// ```?trim={[true]}``` When trim is set to true, the underlying schemas are not returned for entities
//
// ```?apiVersion={apiVersion}``` If apiVersion is unspecified then all models are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?search={componentname}``` If search is non empty then a greedy search is performed
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// ```?annotations={["true"/"false"/]}``` If "true" components having "isAnnotation" property as true are "only" returned, If false all components except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation components being returned.
// responses:
//
//	200: []meshmodelComponentsDuplicateResponseWrapper
func (h *Handler) GetMeshmodelComponentByCategory(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	cat := mux.Vars(r)["category"]
	queryParams := r.URL.Query()
	v := queryParams.Get("version")
	returnAnnotationComp := queryParams.Get("annotations")
	filter := &regv1beta1.ComponentFilter{
		CategoryName: cat,
		Version:      v,
		Trim:         queryParams.Get("trim") == "true",
		APIVersion:   queryParams.Get("apiVersion"),
		Limit:        limit,
		Offset:       offset,
		OrderOn:      order,
		Sort:         sort,
		Annotations:  returnAnnotationComp,
	}
	if search != "" {
		filter.Greedy = true
		filter.DisplayName = search
	}
	entities, count, _, _ := h.registryManager.GetEntities(filter)
	comps := prettifyCompDefSchema(entities)

	var pgSize int64
	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	response := models.MeshmodelComponentsDuplicateAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Components: models.FindDuplicateComponents(comps),
	}

	if err := enc.Encode(response); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route GET /api/meshmodels/components GetAllMeshmodelComponents idGetAllMeshmodelComponents
// Handle GET request for getting meshmodel components across all models and categories
//
// # Components can be further filtered through query parameter
//
// ```?version={version}```
//
// ```?apiVersion={apiVersion}``` If apiVersion is unspecified then all models are returned
//
// ```?order={field}``` orders on the passed field
//
// ```?search={componentname}``` If search is non empty then a greedy search is performed
//
// ```?trim={[true]}``` When trim is set to true, the underlying schemas are not returned for entities
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?page={page-number}``` Default page number is 1
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// ```?annotations={["true"/"false"/]}``` If "true" components having "isAnnotation" property as true are "only" returned, If false all components except "annotations" are returned. Any other value of the query parameter results in both annoations as well as non-annotation components being returned.
// responses:
//  200: meshmodelComponentsDuplicateResponseWrapper

func (h *Handler) GetAllMeshmodelComponents(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)
	queryParams := r.URL.Query()
	v := queryParams.Get("version")
	returnAnnotationComp := queryParams.Get("annotations")
	filter := &regv1beta1.ComponentFilter{
		Version:     v,
		Trim:        queryParams.Get("trim") == "true",
		APIVersion:  queryParams.Get("apiVersion"),
		Limit:       limit,
		Offset:      offset,
		OrderOn:     order,
		Sort:        sort,
		Annotations: returnAnnotationComp,
	}
	if search != "" {
		filter.Greedy = true
		filter.DisplayName = search
	}
	entities, count, _, _ := h.registryManager.GetEntities(filter)
	comps := prettifyCompDefSchema(entities)

	var pgSize int64

	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}

	res := models.MeshmodelComponentsDuplicateAPIResponse{
		Page:       page,
		PageSize:   int(pgSize),
		Count:      count,
		Components: models.FindDuplicateComponents(comps),
	}

	if err := enc.Encode(res); err != nil {
		h.log.Error(ErrGetMeshModels(err)) //TODO: Add appropriate meshkit error
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route POST /api/meshmodel/components/register MeshmodelValidate idPostMeshModelValidate
// Handle POST request for registering meshmodel components.
//
// Validate the given value with the given schema
// responses:
// 	200:

// request body should be json
// request body should be of ComponentCapability format
func (h *Handler) RegisterMeshmodelComponents(rw http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	var cc registry.MeshModelRegistrantData
	err := dec.Decode(&cc)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	var c component.ComponentDefinition
	switch cc.EntityType {
	case entity.ComponentDefinition:
		var isModelError bool
		var isRegistranError bool
		err = json.Unmarshal(cc.Entity, &c)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		utils.WriteSVGsOnFileSystem(&c)
		isRegistranError, isModelError, err = h.registryManager.RegisterEntity(cc.Connection, &c)
		helpers.HandleError(cc.Connection, &c, err, isModelError, isRegistranError)
	}
	err = helpers.WriteLogsToFiles()
	if err != nil {
		h.log.Error(err)
	}
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	go h.config.MeshModelSummaryChannel.Publish()
}

// swagger:route GET /api/meshmodels/registrants GetMeshmodelRegistrants
// Handle GET request for getting all meshmodel registrants
//
// # Returns a list of registrants and summary count of its models, components, and relationships
//
// ```?page={pagenumber}``` Default page number is 1
//
// ```?order={field}``` orders on the passed field
//
// ```?search={Hostname}``` Gets host by the name
//
// ```?sort={[asc/desc]}``` Default behavior is asc
//
// ```?pagesize={pagesize}``` Default pagesize is 25. To return all results: ```pagesize=all```
//
// responses:
//	200: []meshmodelRegistrantsResponseWrapper

func (h *Handler) GetMeshmodelRegistrants(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	page, offset, limit, search, order, sort, _ := getPaginationParams(r)

	filter := &_models.HostFilter{
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		OrderOn: order,
	}
	if search != "" {
		filter.Greedy = true
		filter.DisplayName = search
	}
	hosts, count, err := h.registryManager.GetRegistrants(filter)
	if err != nil {
		h.log.Error(ErrGetMeshModels(err))
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
		return
	}

	var pgSize int64

	if limit == 0 {
		pgSize = count
	} else {
		pgSize = int64(limit)
	}
	res := models.MeshmodelRegistrantsAPIResponse{
		Page:        page,
		PageSize:    int(pgSize),
		Count:       count,
		Registrants: hosts,
	}

	if err := enc.Encode(res); err != nil {
		h.log.Error(ErrGetMeshModels(err))
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}

// swagger:route POST /api/meshmodel/update/status MeshModelUpdateEntityStatus idPostMeshModelUpdateEntityStatus
// Handle POST request for updating the ignore status of a model.
//
// Update the ignore status of a model based on the provided parameters.
//
// responses:
// 	200: noContentWrapper

// request body should be json
// request body should be of struct containing ID and Status fields
func (h *Handler) UpdateEntityStatus(rw http.ResponseWriter, r *http.Request, _ *models.Preference, user *models.User, provider models.Provider) {
	dec := json.NewDecoder(r.Body)
	userID := uuid.FromStringOrNil(user.ID)
	entityType := mux.Vars(r)["entityType"]
	var updateData struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		DisplayName string `json:"displayname"`
	}
	err := dec.Decode(&updateData)
	if err != nil {
		h.log.Error(ErrRequestBody(err))
		http.Error(rw, ErrRequestBody(err).Error(), http.StatusInternalServerError)
		return
	}

	eventBuilder := events.NewEvent().ActedUpon(userID).FromUser(userID).FromSystem(*h.SystemID).WithCategory(entityType).WithAction("update")
	err = h.registryManager.UpdateEntityStatus(updateData.ID, updateData.Status, entityType)
	if err != nil {
		eventBuilder.WithSeverity(events.Error).WithDescription(fmt.Sprintf("Failed to update '%s' status to %s", updateData.DisplayName, updateData.Status)).WithMetadata(map[string]interface{}{
			"error": err,
		})
		_event := eventBuilder.Build()
		_ = provider.PersistEvent(_event)
		go h.config.EventBroadcaster.Publish(userID, _event)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	description := fmt.Sprintf("Status of '%s' updated to %s.", updateData.DisplayName, updateData.Status)

	event := eventBuilder.WithSeverity(events.Informational).WithDescription(description).Build()
	_ = provider.PersistEvent(event)
	go h.config.EventBroadcaster.Publish(userID, event)

	// Respond with success status
	rw.WriteHeader(http.StatusNoContent)
}

func prettifyCompDefSchema(entities []entity.Entity) []component.ComponentDefinition {
	var comps []component.ComponentDefinition
	for _, r := range entities {
		comp, ok := r.(*component.ComponentDefinition)
		if ok {
			m := make(map[string]interface{})
			_ = json.Unmarshal([]byte(comp.Component.Schema), &m)
			m = core.Format.Prettify(m, true)
			b, _ := json.Marshal(m)
			comp.Component.Schema = string(b)
			comps = append(comps, *comp)
		}
	}
	return comps
}

// swagger:route POST /api/meshmodel/register RegisterMeshmodels idRegisterMeshmodels
// Handle POST request for registering entites like components and relationships model.
//
// Register model based on thier Schema Version.
//
// responses:
// 	200: noContentWrapper

// request content byte in form value and header of the type in form

func (h *Handler) RegisterMeshmodels(rw http.ResponseWriter, r *http.Request, _ *models.Preference, user *models.User, provider models.Provider) {
	var response models.RegistryAPIResponse
	regErrorStore := models.NewRegistrationFailureLogHandler()
	var mu sync.Mutex
	defer func() {
		_ = r.Body.Close()
	}()
	userID := uuid.FromStringOrNil(user.ID)
	var message string

	var importRequest struct {
		ImportBody struct {
			ModelFile []byte `json:"model_file"`
			URL       string `json:"url,omitempty"`
			FileName  string `json:"file_name,omitempty"`
		} `json:"importBody"`
		UploadType string `json:"uploadType"`
	}

	err := json.NewDecoder(r.Body).Decode(&importRequest)
	if err != nil {
		h.log.Info("Error in unmarshalling request body")
		http.Error(rw, "Invalid request format", http.StatusBadRequest)
		return
	}
	base64Data, err := json.Marshal(importRequest.ImportBody.ModelFile)
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}
	base64String := string(base64Data)
	// Remove double quotes
	base64String = base64String[1 : len(base64String)-1]

	decodedBytes, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		http.Error(rw, "Invalid base64 data", http.StatusBadRequest)
		return
	}

	switch importRequest.UploadType {
	case "url":
		// Future implementation
		break
	case "file":
		// Create a temporary file and write the data
		tempFile, err := os.CreateTemp("", importRequest.ImportBody.FileName)
		if err != nil {
			err = meshkitutils.ErrCreateFile(err, "Error creating temp file")
			h.handleError(rw, err, err.Error())
			h.sendErrorEvent(userID, provider, "Error creating temp file", err)
			return
		}
		defer os.Remove(tempFile.Name())

		_, err = tempFile.Write(decodedBytes)
		if err != nil {
			err = meshkitutils.ErrWriteFile(err, importRequest.ImportBody.FileName)
			h.handleError(rw, err, err.Error())
			h.sendErrorEvent(userID, provider, "Error writing to temp file", err)
			return
		}
		tempFile.Close()
		// Pass the temp file path to the RegistrationHelper
		registrationHelper := registration.NewRegistrationHelper(
			utils.UI,
			h.registryManager,
			regErrorStore,
		)

		dir := registration.NewDir(tempFile.Name())
		registrationHelper.Register(dir)
		for _, pkg := range registrationHelper.PkgUnits {
			registeredModel := pkg.Model
			for _, comp := range pkg.Components {
				incrementCountersOnSuccess(&mu, entity.ComponentDefinition, &response.EntityCount.CompCount, &response.EntityCount.RelCount, &response.EntityCount.ModelCount)
				componentBytes, _ := json.Marshal(comp)
				addSuccessfulEntry(componentBytes, entity.ComponentDefinition, &response)
			}
			for _, rel := range pkg.Relationships {
				incrementCountersOnSuccess(&mu, entity.ComponentDefinition, &response.EntityCount.CompCount, &response.EntityCount.RelCount, &response.EntityCount.ModelCount)
				relationshipBytes, _ := json.Marshal(rel)
				addSuccessfulEntry(relationshipBytes, entity.ComponentDefinition, &response)
			}
			modelBytes, _ := json.Marshal(registeredModel)
			incrementCountersOnSuccess(&mu, entity.Model, &response.EntityCount.CompCount, &response.EntityCount.RelCount, &response.EntityCount.ModelCount)
			addSuccessfulEntry(modelBytes, entity.Model, &response)

		}

		if regErrorStore != nil {
			// Process InvalidDefinitions
			for _, errEntry := range regErrorStore.GetEntityRegErrors() {
				// Increment error counters
				if errEntry.EntityType == "Unknown" {
					errEntry.EntityType = ""
				}
				incrementCountersOnErr(&mu, errEntry.EntityType, &response)
				// Add unsuccessful entry
				path := errEntry.EntityName
				err := errEntry.Err
				entityTypeStr := string(errEntry.EntityType)
				addUnsuccessfulEntry(path, &response, err, entityTypeStr)
			}
		}
		var errMsg string
		// Construct the response message
		message = writeMessageString(&response)
		if response.EntityCount.TotalErrCount > 0 {
			errMsg = ErrMsgContruct(&response)
		}

		h.sendSuccessResponse(rw, userID, provider, message, errMsg, &response)
	}
}

// swagger:route POST /api/meshmodel/export ExportModel idExportModel
// Handle GET request for exporting a model.
//
// # Export model with the given id in the output format specified
//
// ```?id={id}```
// ```?output_format={output_format}``` Can be `json`, `yaml`, or `oci`. Default is `oci`
//
// responses:
//
//	200: []byte
func (h *Handler) ExportModel(rw http.ResponseWriter, r *http.Request) {
	modelId := r.URL.Query().Get("id")

	// 1. Get the model data
	modelFilter := &regv1beta1.ModelFilter{
		Id:            modelId,
		Components:    true,
		Relationships: true,
	}
	e, _, _, err := h.registryManager.GetEntities(modelFilter)
	if err != nil {
		h.log.Error(ErrGetMeshModels(err))
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
		return // Ensure the function returns after handling the error
	}

	model := e[0].(*model.ModelDefinition)

	// 2. Convert it to oci
	temp := os.TempDir()
	modelDir := filepath.Join(temp, model.Name)
	err = os.Mkdir(modelDir, 0700)
	if err != nil {
		err = meshkitutils.ErrCreateDir(err, "Error creating temp directory")
		h.log.Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(modelDir)
	_ = utils.ReplaceSVGData(model)

	err = model.WriteModelDefinition(filepath.Join(modelDir, "model.json"), "json")
	if err != nil {
		h.log.Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return // Ensure the function returns after handling the error
	}

	// build oci image for the model
	img, err := oci.BuildImage(modelDir)
	if err != nil {
		h.log.Error(err) // TODO: Add appropriate meshkit error
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return // Ensure the function returns after handling the error
	}
	tarfileName := filepath.Join(modelDir, "model.tar")
	err = oci.SaveOCIArtifact(img, tarfileName, model.Name)
	if err != nil {
		h.log.Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return // Ensure the function returns after handling the error
	}
	// 3. Send response
	byt, _ := os.ReadFile(tarfileName)
	rw.Header().Add("Content-Type", "application/x-tar")
	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.tar\"", model.Name))
	rw.Header().Set("Content-Length", fmt.Sprintf("%d", len(byt))) // Correctly format the Content-Length header
	_, err = rw.Write(byt)
	if err != nil {
		h.log.Error(ErrGetMeshModels(err))
		http.Error(rw, ErrGetMeshModels(err).Error(), http.StatusInternalServerError)
	}
}
func RegisterEntity(content []byte, entityType entity.EntityType, h *Handler) error {
	switch entityType {
	case entity.ComponentDefinition:
		var c component.ComponentDefinition
		err := json.Unmarshal(content, &c)
		if err != nil {
			return meshkitutils.ErrUnmarshal(err)
		}
		isRegistrantError, isModelError, err := h.registryManager.RegisterEntity(connection.Connection{
			Kind: c.Model.Registrant.Kind,
		}, &c)
		helpers.HandleError(connection.Connection{
			Kind: c.Model.Registrant.Kind,
		}, &c, err, isModelError, isRegistrantError)
		return nil
	case entity.RelationshipDefinition:
		var r relationship.RelationshipDefinition
		err := json.Unmarshal(content, &r)
		if err != nil {
			return meshkitutils.ErrUnmarshal(err)
		}
		isRegistrantError, isModelError, err := h.registryManager.RegisterEntity(connection.Connection{
			Kind: r.Model.Registrant.Kind,
		}, &r)
		helpers.HandleError(connection.Connection{
			Kind: r.Model.Registrant.Kind,
		}, &r, err, isModelError, isRegistrantError)
		return nil
	}
	return meshkitutils.ErrInvalidSchemaVersion
}
