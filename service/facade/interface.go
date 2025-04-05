package facade

import (
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/deity_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/prarthana_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/shlok_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/stotra_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
)

type Service interface {
	ShlokIngestionService() shlok_ingestion.Service
	StotraIngestionService() stotra_ingestion.Service
	PrarthanaIngestionService() prarthana_ingestion.Service
	DeityIngestionService() deity_ingestion.Service
	ZohoAuthService() zoho.Service
}
