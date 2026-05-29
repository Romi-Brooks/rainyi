package repository

import (
	"rain-yi-backend/config"
	"rain-yi-backend/model"
)

type PersonaRepository struct{}

func NewPersonaRepository() *PersonaRepository {
	return &PersonaRepository{}
}

func (r *PersonaRepository) Create(persona *model.Persona) error {
	return config.DB.Create(persona).Error
}

func (r *PersonaRepository) FindByID(id int64) (*model.Persona, error) {
	var persona model.Persona
	err := config.DB.First(&persona, id).Error
	if err != nil {
		return nil, err
	}
	return &persona, nil
}

func (r *PersonaRepository) FindByUserID(userID int64) ([]model.Persona, error) {
	var personas []model.Persona
	err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&personas).Error
	return personas, err
}

func (r *PersonaRepository) FindAllVisible(userID int64) ([]model.Persona, error) {
	var personas []model.Persona
	err := config.DB.Where("user_id = ? OR user_id = 0", userID).
		Order("user_id ASC, created_at DESC").
		Find(&personas).Error
	return personas, err
}

func (r *PersonaRepository) Update(persona *model.Persona) error {
	return config.DB.Save(persona).Error
}

func (r *PersonaRepository) Delete(id int64) error {
	r.DeleteSkillNodesByPersonaID(id)
	return config.DB.Delete(&model.Persona{}, id).Error
}

func (r *PersonaRepository) FindByName(name string) (*model.Persona, error) {
	var persona model.Persona
	err := config.DB.Where("name = ?", name).First(&persona).Error
	if err != nil {
		return nil, err
	}
	return &persona, nil
}

func (r *PersonaRepository) DeleteAllByUserID(userID int64) error {
	personas, err := r.FindByUserID(userID)
	if err != nil {
		return err
	}
	for _, p := range personas {
		r.DeleteSkillNodesByPersonaID(p.ID)
	}
	return config.DB.Where("user_id = ?", userID).Delete(&model.Persona{}).Error
}

func (r *PersonaRepository) SetConversationPersona(convID int64, personaID *int64) error {
	return config.DB.Model(&model.Conversation{}).Where("id = ?", convID).Update("persona_id", personaID).Error
}

func (r *PersonaRepository) CountSkillNodesByPersonaID(personaID int64) (int64, error) {
	var count int64
	err := config.DB.Model(&model.SkillNode{}).Where("persona_id = ?", personaID).Count(&count).Error
	return count, err
}

func (r *PersonaRepository) CreateSkillNode(sn *model.SkillNode) error {
	return config.DB.Create(sn).Error
}

func (r *PersonaRepository) CreateSkillNodesBatch(nodes []model.SkillNode) error {
	if len(nodes) == 0 {
		return nil
	}
	return config.DB.Create(&nodes).Error
}

func (r *PersonaRepository) FindSkillNodesByPersonaID(personaID int64) ([]model.SkillNode, error) {
	var nodes []model.SkillNode
	err := config.DB.Where("persona_id = ?", personaID).
		Order("priority ASC, created_at ASC").
		Find(&nodes).Error
	return nodes, err
}

func (r *PersonaRepository) FindSkillNodeByID(id int64) (*model.SkillNode, error) {
	var sn model.SkillNode
	err := config.DB.First(&sn, id).Error
	if err != nil {
		return nil, err
	}
	return &sn, nil
}

func (r *PersonaRepository) DeleteSkillNode(id int64) error {
	r.DeleteSkillKVsBySkillNodeID(id)
	return config.DB.Delete(&model.SkillNode{}, id).Error
}

func (r *PersonaRepository) DeleteSkillNodesByPersonaID(personaID int64) error {
	var nodes []model.SkillNode
	config.DB.Where("persona_id = ?", personaID).Find(&nodes)
	for _, n := range nodes {
		r.DeleteSkillKVsBySkillNodeID(n.ID)
	}
	return config.DB.Where("persona_id = ?", personaID).Delete(&model.SkillNode{}).Error
}

func (r *PersonaRepository) CreateSkillKVsBatch(kvs []model.SkillKV) error {
	if len(kvs) == 0 {
		return nil
	}
	return config.DB.Create(&kvs).Error
}

func (r *PersonaRepository) FindSkillKVsBySkillNodeID(skillNodeID int64) ([]model.SkillKV, error) {
	var kvs []model.SkillKV
	err := config.DB.Where("skill_node_id = ?", skillNodeID).
		Order("sort_order ASC").
		Find(&kvs).Error
	return kvs, err
}

func (r *PersonaRepository) DeleteSkillKVsBySkillNodeID(skillNodeID int64) error {
	return config.DB.Where("skill_node_id = ?", skillNodeID).Delete(&model.SkillKV{}).Error
}
