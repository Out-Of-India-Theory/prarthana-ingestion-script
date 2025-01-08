package facade

import (
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/deity_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/prarthana_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/shlok_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/stotra_ingestion"
)

type Service interface {
	ShlokIngestionService() shlok_ingestion.Service
	StotraIngestionService() stotra_ingestion.Service
	PrarthanaIngestionService() prarthana_ingestion.Service
	DeityIngestionService() deity_ingestion.Service
}
