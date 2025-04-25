package repository

import (
	"server/internal/domain"
	"encoding/xml"
	"os"
	"sync"
)

type XMLRepository struct {
	filePath string
	mu       sync.Mutex
}

func NewXMLRepository(filePath string) *XMLRepository {
	return &XMLRepository{
		filePath: filePath,
	}
}

func (r *XMLRepository) SaveGameResult(game *domain.Game) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    // Создаем файл если не существует
    file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    // Добавляем XML заголовок для новых файлов
    if stat, _ := file.Stat(); stat.Size() == 0 {
        if _, err := file.WriteString(xml.Header); err != nil {
            return err
        }
    }

    encoder := xml.NewEncoder(file)
    encoder.Indent("", "  ")
    if err := encoder.Encode(game.ToGameResult()); err != nil {
        return err
    }

    // Добавляем перенос строки между записями
    if _, err := file.WriteString("\n"); err != nil {
        return err
    }

    return nil
}