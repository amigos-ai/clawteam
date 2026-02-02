package instance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Registry struct {
	dir  string
	path string
}

type registryData struct {
	Instances map[string]*Instance `json:"instances"`
}

func NewRegistry(dir string) *Registry {
	return &Registry{
		dir:  dir,
		path: filepath.Join(dir, "instances.json"),
	}
}

func (r *Registry) load() (*registryData, error) {
	data, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return &registryData{Instances: make(map[string]*Instance)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read registry: %w", err)
	}

	var reg registryData
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("unmarshal registry: %w", err)
	}
	if reg.Instances == nil {
		reg.Instances = make(map[string]*Instance)
	}

	return &reg, nil
}

func (r *Registry) save(data *registryData) error {
	if err := os.MkdirAll(r.dir, 0755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(r.path, bytes, 0644); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}

	return nil
}

func (r *Registry) Save(inst *Instance) error {
	data, err := r.load()
	if err != nil {
		return err
	}

	data.Instances[inst.Name] = inst
	return r.save(data)
}

func (r *Registry) Get(name string) (*Instance, error) {
	data, err := r.load()
	if err != nil {
		return nil, err
	}

	inst, ok := data.Instances[name]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", name)
	}

	return inst, nil
}

func (r *Registry) List() ([]*Instance, error) {
	data, err := r.load()
	if err != nil {
		return nil, err
	}

	var instances []*Instance
	for _, inst := range data.Instances {
		instances = append(instances, inst)
	}

	return instances, nil
}

func (r *Registry) Delete(name string) error {
	data, err := r.load()
	if err != nil {
		return err
	}

	delete(data.Instances, name)
	return r.save(data)
}
