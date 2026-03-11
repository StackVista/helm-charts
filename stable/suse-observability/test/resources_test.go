package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type expectedResources struct {
	memoryLimit   string
	cpuLimit      string
	memoryRequest string
	cpuRequest    string
}

// TestGlobalSizingUserResourceOverrides tests that user-specified resource overrides take precedence over sizing profile defaults
func TestResources(t *testing.T) {
	expectedDeployments := map[string]map[string]expectedResources{
		"10-nonha": {
			"suse-observability-correlate":                         {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "1250Mi", memoryLimit: "1250Mi"},
			"suse-observability-e2es":                              {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-hbase-console":                     {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "50Mi", memoryLimit: "50Mi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver":                          {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "1000Mi", memoryLimit: "1000Mi"},
			"suse-observability-router":                            {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "128Mi", memoryLimit: "128Mi"},
			"suse-observability-server":                            {cpuRequest: "1500m", cpuLimit: "3000m", memoryRequest: "5Gi", memoryLimit: "5Gi"},
			"suse-observability-ui":                                {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"20-nonha": {
			"suse-observability-correlate":                         {cpuRequest: "1500m", cpuLimit: "3000m", memoryRequest: "1750Mi", memoryLimit: "1750Mi"},
			"suse-observability-e2es":                              {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-hbase-console":                     {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "50Mi", memoryLimit: "50Mi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver":                          {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "1750Mi", memoryLimit: "1750Mi"},
			"suse-observability-router":                            {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "128Mi", memoryLimit: "128Mi"},
			"suse-observability-server":                            {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "6Gi", memoryLimit: "6Gi"},
			"suse-observability-ui":                                {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"50-nonha": {
			"suse-observability-correlate":                         {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "2000Mi", memoryLimit: "2000Mi"},
			"suse-observability-e2es":                              {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-hbase-console":                     {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "50Mi", memoryLimit: "50Mi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver":                          {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "2250Mi", memoryLimit: "2250Mi"},
			"suse-observability-router":                            {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "128Mi", memoryLimit: "128Mi"},
			"suse-observability-server":                            {cpuRequest: "2500m", cpuLimit: "5000m", memoryRequest: "6500Mi", memoryLimit: "6500Mi"},
			"suse-observability-ui":                                {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"100-nonha": {
			"suse-observability-correlate":                         {cpuRequest: "5000m", cpuLimit: "10000m", memoryRequest: "4000Mi", memoryLimit: "4000Mi"},
			"suse-observability-e2es":                              {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-hbase-console":                     {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "100Mi", memoryLimit: "100Mi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver":                          {cpuRequest: "5500m", cpuLimit: "11000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-router":                            {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "128Mi", memoryLimit: "128Mi"},
			"suse-observability-server":                            {cpuRequest: "4000m", cpuLimit: "8000m", memoryRequest: "8Gi", memoryLimit: "8Gi"},
			"suse-observability-ui":                                {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"150-ha": {
			"suse-observability-api":                               {cpuRequest: "1500m", cpuLimit: "3000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-authorization-sync":                {cpuRequest: "250m", cpuLimit: "1500m", memoryRequest: "512Mi", memoryLimit: "1Gi"},
			"suse-observability-checks":                            {cpuRequest: "3000m", cpuLimit: "6000m", memoryRequest: "5Gi", memoryLimit: "5Gi"},
			"suse-observability-correlate":                         {cpuRequest: "3", cpuLimit: "6", memoryRequest: "3500Mi", memoryLimit: "3500Mi"},
			"suse-observability-e2es":                              {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-hbase-console":                     {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-health-sync":                       {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "5Gi", memoryLimit: "5Gi"},
			"suse-observability-initializer":                       {cpuRequest: "50m", cpuLimit: "1500m", memoryRequest: "512Mi", memoryLimit: "1500Mi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-notification":                      {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "3000Mi", memoryLimit: "3000Mi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "100Mi", memoryLimit: "100Mi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver-base":                     {cpuRequest: "4000m", cpuLimit: "8000m", memoryRequest: "5500Mi", memoryLimit: "5500Mi"},
			"suse-observability-receiver-logs":                     {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-receiver-process-agent":            {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "2500Mi", memoryLimit: "2500Mi"},
			"suse-observability-router":                            {cpuRequest: "120m", cpuLimit: "240m", memoryRequest: "128Mi", memoryLimit: "128Mi"},
			"suse-observability-slicing":                           {cpuRequest: "300m", cpuLimit: "1500m", memoryRequest: "1800Mi", memoryLimit: "2000Mi"},
			"suse-observability-state":                             {cpuRequest: "2", cpuLimit: "4", memoryRequest: "4000Mi", memoryLimit: "4000Mi"},
			"suse-observability-sync":                              {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "6500Mi", memoryLimit: "6500Mi"},
			"suse-observability-ui":                                {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"250-ha": {
			"suse-observability-api":                               {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "5Gi", memoryLimit: "5Gi"},
			"suse-observability-authorization-sync":                {cpuRequest: "250m", cpuLimit: "1500m", memoryRequest: "512Mi", memoryLimit: "1Gi"},
			"suse-observability-checks":                            {cpuRequest: "3000m", cpuLimit: "6000m", memoryRequest: "5Gi", memoryLimit: "5Gi"},
			"suse-observability-correlate":                         {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-e2es":                              {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-hbase-console":                     {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-health-sync":                       {cpuRequest: "1500m", cpuLimit: "3000m", memoryRequest: "6Gi", memoryLimit: "6Gi"},
			"suse-observability-initializer":                       {cpuRequest: "50m", cpuLimit: "1500m", memoryRequest: "512Mi", memoryLimit: "1500Mi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-notification":                      {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "3000Mi", memoryLimit: "3000Mi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "100Mi", memoryLimit: "100Mi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver-base":                     {cpuRequest: "4000m", cpuLimit: "8000m", memoryRequest: "6500Mi", memoryLimit: "6500Mi"},
			"suse-observability-receiver-logs":                     {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-receiver-process-agent":            {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-router":                            {cpuRequest: "120m", cpuLimit: "240m", memoryRequest: "128Mi", memoryLimit: "128Mi"},
			"suse-observability-slicing":                           {cpuRequest: "300m", cpuLimit: "600m", memoryRequest: "1500Mi", memoryLimit: "1500Mi"},
			"suse-observability-state":                             {cpuRequest: "2", cpuLimit: "4", memoryRequest: "4000Mi", memoryLimit: "4000Mi"},
			"suse-observability-sync":                              {cpuRequest: "3000m", cpuLimit: "6000m", memoryRequest: "8Gi", memoryLimit: "8Gi"},
			"suse-observability-ui":                                {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"500-ha": {
			"suse-observability-api":                               {cpuRequest: "3000m", cpuLimit: "6000m", memoryRequest: "6Gi", memoryLimit: "6Gi"},
			"suse-observability-authorization-sync":                {cpuRequest: "250m", cpuLimit: "1500m", memoryRequest: "512Mi", memoryLimit: "1Gi"},
			"suse-observability-checks":                            {cpuRequest: "3000m", cpuLimit: "6000m", memoryRequest: "5Gi", memoryLimit: "5Gi"},
			"suse-observability-correlate":                         {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "3000Mi", memoryLimit: "3000Mi"},
			"suse-observability-e2es":                              {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-hbase-console":                     {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-health-sync":                       {cpuRequest: "4000m", cpuLimit: "8000m", memoryRequest: "8Gi", memoryLimit: "8Gi"},
			"suse-observability-initializer":                       {cpuRequest: "50m", cpuLimit: "1500m", memoryRequest: "512Mi", memoryLimit: "1500Mi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-notification":                      {cpuRequest: "1500m", cpuLimit: "3000m", memoryRequest: "3000Mi", memoryLimit: "3000Mi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "100Mi", memoryLimit: "100Mi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver-base":                     {cpuRequest: "6000m", cpuLimit: "12000m", memoryRequest: "7Gi", memoryLimit: "7Gi"},
			"suse-observability-receiver-logs":                     {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-receiver-process-agent":            {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-router":                            {cpuRequest: "120m", cpuLimit: "240m", memoryRequest: "128Mi", memoryLimit: "128Mi"},
			"suse-observability-slicing":                           {cpuRequest: "300m", cpuLimit: "1500m", memoryRequest: "1800Mi", memoryLimit: "2000Mi"},
			"suse-observability-state":                             {cpuRequest: "2", cpuLimit: "4", memoryRequest: "4000Mi", memoryLimit: "4000Mi"},
			"suse-observability-sync":                              {cpuRequest: "4000m", cpuLimit: "8000m", memoryRequest: "8Gi", memoryLimit: "8Gi"},
			"suse-observability-ui":                                {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"4000-ha": {
			"suse-observability-api":                               {cpuRequest: "7000m", cpuLimit: "9000m", memoryRequest: "10Gi", memoryLimit: "12Gi"},
			"suse-observability-authorization-sync":                {cpuRequest: "250m", cpuLimit: "1500m", memoryRequest: "512Mi", memoryLimit: "1Gi"},
			"suse-observability-checks":                            {cpuRequest: "6000m", cpuLimit: "7000m", memoryRequest: "5Gi", memoryLimit: "5Gi"},
			"suse-observability-correlate":                         {cpuRequest: "5000m", cpuLimit: "6000m", memoryRequest: "4000Mi", memoryLimit: "6000Mi"},
			"suse-observability-e2es":                              {cpuRequest: "2000m", cpuLimit: "3000m", memoryRequest: "768Mi", memoryLimit: "1024Mi"},
			"suse-observability-hbase-console":                     {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-health-sync":                       {cpuRequest: "5000m", cpuLimit: "7000m", memoryRequest: "12Gi", memoryLimit: "14Gi"},
			"suse-observability-initializer":                       {cpuRequest: "1000m", cpuLimit: "1500m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-notification":                      {cpuRequest: "2000m", cpuLimit: "3000m", memoryRequest: "3200Mi", memoryLimit: "4500Mi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "100Mi", memoryLimit: "100Mi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver-base":                     {cpuRequest: "8", cpuLimit: "9000m", memoryRequest: "6Gi", memoryLimit: "7Gi"},
			"suse-observability-receiver-logs":                     {cpuRequest: "2", cpuLimit: "3", memoryRequest: "4Gi", memoryLimit: "6Gi"},
			"suse-observability-receiver-process-agent":            {cpuRequest: "6", cpuLimit: "8000m", memoryRequest: "4Gi", memoryLimit: "6Gi"},
			"suse-observability-router":                            {cpuRequest: "2000m", cpuLimit: "2500m", memoryRequest: "256Mi", memoryLimit: "512Mi"},
			"suse-observability-slicing":                           {cpuRequest: "1000m", cpuLimit: "1500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-state":                             {cpuRequest: "3", cpuLimit: "5", memoryRequest: "4000Mi", memoryLimit: "6000Mi"},
			"suse-observability-sync":                              {cpuRequest: "12", cpuLimit: "14", memoryRequest: "14000Mi", memoryLimit: "16000Mi"},
			"suse-observability-ui":                                {cpuRequest: "100m", cpuLimit: "500m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"default-nonha": {
			"suse-observability-anomaly-detection-spotlight-manager": {cpuRequest: "1", cpuLimit: "1", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-anomaly-detection-spotlight-worker":  {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-correlate":                           {cpuRequest: "600m", cpuLimit: "2000m", memoryRequest: "2800Mi", memoryLimit: "2800Mi"},
			"suse-observability-e2es":                                {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "768Mi", memoryLimit: "1500Mi"},
			"suse-observability-hbase-console":                       {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-kafkaup-operator-kafkaup":            {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-minio":                               {cpuRequest: "500m", cpuLimit: "1", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-prometheus-elasticsearch-exporter":   {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "100Mi", memoryLimit: "100Mi"},
			"suse-observability-rbac-agent":                          {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver":                            {cpuRequest: "1000m", cpuLimit: "3000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-router":                              {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "128Mi", memoryLimit: "128Mi"},
			"suse-observability-server":                              {cpuRequest: "3600m", cpuLimit: "3600m", memoryRequest: "8Gi", memoryLimit: "8Gi"},
			"suse-observability-ui":                                  {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"default-ha": {
			"suse-observability-anomaly-detection-spotlight-manager": {cpuRequest: "1", cpuLimit: "1", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-anomaly-detection-spotlight-worker":  {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-api":                                 {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "2Gi", memoryLimit: "2Gi"},
			"suse-observability-authorization-sync":                  {cpuRequest: "250m", cpuLimit: "1500m", memoryRequest: "512Mi", memoryLimit: "1Gi"},
			"suse-observability-checks":                              {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "4000Mi", memoryLimit: "4000Mi"},
			"suse-observability-correlate":                           {cpuRequest: "600m", cpuLimit: "2000m", memoryRequest: "2800Mi", memoryLimit: "2800Mi"},
			"suse-observability-e2es":                                {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "768Mi", memoryLimit: "1500Mi"},
			"suse-observability-hbase-console":                       {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-health-sync":                         {cpuRequest: "400m", cpuLimit: "1500m", memoryRequest: "3500Mi", memoryLimit: "3500Mi"},
			"suse-observability-initializer":                         {cpuRequest: "250m", cpuLimit: "1500m", memoryRequest: "512Mi", memoryLimit: "1500Mi"},
			"suse-observability-kafkaup-operator-kafkaup":            {cpuRequest: "25m", cpuLimit: "50m", memoryRequest: "32Mi", memoryLimit: "64Mi"},
			"suse-observability-minio":                               {cpuRequest: "500m", cpuLimit: "1", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-notification":                        {cpuRequest: "250m", cpuLimit: "750m", memoryRequest: "1500Mi", memoryLimit: "1500Mi"},
			"suse-observability-prometheus-elasticsearch-exporter":   {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "100Mi", memoryLimit: "100Mi"},
			"suse-observability-rbac-agent":                          {cpuRequest: "", cpuLimit: "", memoryRequest: "25Mi", memoryLimit: "40Mi"},
			"suse-observability-receiver-base":                       {cpuRequest: "1000m", cpuLimit: "3000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-receiver-logs":                       {cpuRequest: "1000m", cpuLimit: "3000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-receiver-process-agent":              {cpuRequest: "1000m", cpuLimit: "3000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-router":                              {cpuRequest: "100m", cpuLimit: "100m", memoryRequest: "128Mi", memoryLimit: "128Mi"},
			"suse-observability-slicing":                             {cpuRequest: "250m", cpuLimit: "1500m", memoryRequest: "1800Mi", memoryLimit: "2000Mi"},
			"suse-observability-state":                               {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "1536Mi", memoryLimit: "2000Mi"},
			"suse-observability-sync":                                {cpuRequest: "750m", cpuLimit: "3000m", memoryRequest: "3Gi", memoryLimit: "4Gi"},
			"suse-observability-ui":                                  {cpuRequest: "50m", cpuLimit: "50m", memoryRequest: "64Mi", memoryLimit: "64Mi"},
		},
		"10-nonha-overrides": {
			"suse-observability-correlate":                         {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-e2es":                              {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-hbase-console":                     {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-receiver":                          {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-router":                            {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-server":                            {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-ui":                                {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
		},
		"150-ha-overrides": {
			"suse-observability-api":                               {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-authorization-sync":                {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-checks":                            {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-correlate":                         {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-e2es":                              {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-hbase-console":                     {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-health-sync":                       {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-initializer":                       {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-notification":                      {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-receiver-base":                     {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-receiver-logs":                     {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-receiver-process-agent":            {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-router":                            {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-slicing":                           {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-state":                             {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-sync":                              {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-ui":                                {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"}},
		"4000-ha-overrides": {
			"suse-observability-api":                               {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-authorization-sync":                {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-checks":                            {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-correlate":                         {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-e2es":                              {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-hbase-console":                     {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-health-sync":                       {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-initializer":                       {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-notification":                      {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-receiver-base":                     {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-receiver-logs":                     {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-receiver-process-agent":            {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-router":                            {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-slicing":                           {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-state":                             {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-sync":                              {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-ui":                                {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
		},
		"10-nonha-global-overrides": {
			"suse-observability-correlate":                         {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-e2es":                              {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-hbase-console":                     {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-receiver":                          {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-router":                            {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-server":                            {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-ui":                                {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
		},
		"150-ha-global-overrides": {
			"suse-observability-api":                               {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-authorization-sync":                {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-checks":                            {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-correlate":                         {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-e2es":                              {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-hbase-console":                     {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-health-sync":                       {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-initializer":                       {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-notification":                      {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-receiver-base":                     {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-receiver-logs":                     {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-receiver-process-agent":            {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-router":                            {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-slicing":                           {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-state":                             {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-sync":                              {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-ui":                                {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
		},
		"4000-ha-global-overrides": {
			"suse-observability-api":                               {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-authorization-sync":                {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-checks":                            {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-correlate":                         {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-e2es":                              {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-hbase-console":                     {cpuRequest: "888", cpuLimit: "888", memoryRequest: "888Gi", memoryLimit: "888Gi"},
			"suse-observability-health-sync":                       {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-initializer":                       {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-kafkaup-operator-kafkaup":          {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-notification":                      {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-prometheus-elasticsearch-exporter": {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-rbac-agent":                        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-receiver-base":                     {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-receiver-logs":                     {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-receiver-process-agent":            {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-router":                            {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-slicing":                           {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-state":                             {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-sync":                              {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-ui":                                {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
		},
	}
	expectedStatefulsets := map[string]map[string]expectedResources{
		"10-nonha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "2500Mi", memoryLimit: "2500Mi"},
			"suse-observability-hbase-stackgraph":     {cpuRequest: "500m", cpuLimit: "1500m", memoryRequest: "2250Mi", memoryLimit: "2250Mi"},
			"suse-observability-hbase-tephra-mono":    {cpuRequest: "50m", cpuLimit: "100m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-kafka":                {cpuRequest: "800m", cpuLimit: "1600m", memoryRequest: "2048Mi", memoryLimit: "2048Mi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "1500Mi", memoryLimit: "1750Mi"},
			"suse-observability-vmagent":              {cpuRequest: "200m", cpuLimit: "200m", memoryRequest: "384Mi", memoryLimit: "640Mi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "100m", cpuLimit: "250m", memoryRequest: "640Mi", memoryLimit: "640Mi"},
		},
		"20-nonha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "3000Mi", memoryLimit: "3000Mi"},
			"suse-observability-hbase-stackgraph":     {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "2500Mi", memoryLimit: "2500Mi"},
			"suse-observability-hbase-tephra-mono":    {cpuRequest: "50m", cpuLimit: "100m", memoryRequest: "600Mi", memoryLimit: "600Mi"},
			"suse-observability-kafka":                {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "2048Mi", memoryLimit: "2048Mi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "2Gi", memoryLimit: "2500Mi"},
			"suse-observability-vmagent":              {cpuRequest: "300m", cpuLimit: "600m", memoryRequest: "600Mi", memoryLimit: "768Mi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "100m", cpuLimit: "250m", memoryRequest: "640Mi", memoryLimit: "640Mi"},
		},
		"50-nonha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-stackgraph":     {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-hbase-tephra-mono":    {cpuRequest: "50m", cpuLimit: "100m", memoryRequest: "750Mi", memoryLimit: "750Mi"},
			"suse-observability-kafka":                {cpuRequest: "1500m", cpuLimit: "3000m", memoryRequest: "2048Mi", memoryLimit: "2048Mi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "1", cpuLimit: "2", memoryRequest: "3500Mi", memoryLimit: "3500Mi"},
			"suse-observability-vmagent":              {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "750Mi", memoryLimit: "750Mi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "100m", cpuLimit: "250m", memoryRequest: "640Mi", memoryLimit: "640Mi"},
		},
		"100-nonha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "750m", cpuLimit: "1500m", memoryRequest: "7Gi", memoryLimit: "7Gi"},
			"suse-observability-hbase-stackgraph":     {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "4500Mi", memoryLimit: "4500Mi"},
			"suse-observability-hbase-tephra-mono":    {cpuRequest: "100m", cpuLimit: "200m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-kafka":                {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "3000Mi", memoryLimit: "3000Mi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "1500m", cpuLimit: "3000m", memoryRequest: "8Gi", memoryLimit: "8Gi"},
			"suse-observability-vmagent":              {cpuRequest: "1250m", cpuLimit: "2500m", memoryRequest: "1250Mi", memoryLimit: "1250Mi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "100m", cpuLimit: "250m", memoryRequest: "640Mi", memoryLimit: "640Mi"},
		},
		"150-ha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hbase-master":   {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-hbase-rs":       {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-hbase-hdfs-dn":        {cpuRequest: "600m", cpuLimit: "1200m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hdfs-nn":        {cpuRequest: "200m", cpuLimit: "400m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-hdfs-snn":       {cpuRequest: "10m", cpuLimit: "500m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-tephra":         {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-kafka":                {cpuRequest: "1", cpuLimit: "2", memoryRequest: "3Gi", memoryLimit: "4Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "2", cpuLimit: "4", memoryRequest: "9Gi", memoryLimit: "9Gi"},
			"suse-observability-victoria-metrics-1":   {cpuRequest: "2", cpuLimit: "4", memoryRequest: "9Gi", memoryLimit: "9Gi"},
			"suse-observability-vmagent":              {cpuRequest: "1500m", cpuLimit: "3000m", memoryRequest: "1500Mi", memoryLimit: "1500Mi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "200m", cpuLimit: "500m", memoryRequest: "640Mi", memoryLimit: "640Mi"},
		},
		"250-ha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hbase-master":   {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-hbase-rs":       {cpuRequest: "3000m", cpuLimit: "6000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hdfs-dn":        {cpuRequest: "600m", cpuLimit: "1200m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hdfs-nn":        {cpuRequest: "200m", cpuLimit: "400m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-hdfs-snn":       {cpuRequest: "10m", cpuLimit: "500m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-tephra":         {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-kafka":                {cpuRequest: "2", cpuLimit: "4", memoryRequest: "3Gi", memoryLimit: "4Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "3", cpuLimit: "6", memoryRequest: "10Gi", memoryLimit: "10Gi"},
			"suse-observability-victoria-metrics-1":   {cpuRequest: "3", cpuLimit: "6", memoryRequest: "10Gi", memoryLimit: "10Gi"},
			"suse-observability-vmagent":              {cpuRequest: "2", cpuLimit: "4", memoryRequest: "2Gi", memoryLimit: "2Gi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "200m", cpuLimit: "500m", memoryRequest: "640Mi", memoryLimit: "640Mi"},
		},
		"500-ha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hbase-master":   {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-hbase-rs":       {cpuRequest: "4000m", cpuLimit: "8000m", memoryRequest: "6Gi", memoryLimit: "6Gi"},
			"suse-observability-hbase-hdfs-dn":        {cpuRequest: "600m", cpuLimit: "1200m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hdfs-nn":        {cpuRequest: "200m", cpuLimit: "400m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-hdfs-snn":       {cpuRequest: "10m", cpuLimit: "500m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-tephra":         {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-kafka":                {cpuRequest: "3", cpuLimit: "6", memoryRequest: "3Gi", memoryLimit: "4Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "3", cpuLimit: "6", memoryRequest: "10Gi", memoryLimit: "10Gi"},
			"suse-observability-victoria-metrics-1":   {cpuRequest: "3", cpuLimit: "6", memoryRequest: "10Gi", memoryLimit: "10Gi"},
			"suse-observability-vmagent":              {cpuRequest: "2500m", cpuLimit: "5000m", memoryRequest: "2Gi", memoryLimit: "2Gi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "200m", cpuLimit: "500m", memoryRequest: "640Mi", memoryLimit: "640Mi"},
		},
		"4000-ha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "4", cpuLimit: "6", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hbase-master":   {cpuRequest: "2000m", cpuLimit: "2500m", memoryRequest: "1024Mi", memoryLimit: "1024Mi"},
			"suse-observability-hbase-hbase-rs":       {cpuRequest: "6000m", cpuLimit: "8000m", memoryRequest: "10Gi", memoryLimit: "12Gi"},
			"suse-observability-hbase-hdfs-dn":        {cpuRequest: "4000m", cpuLimit: "6000m", memoryRequest: "3Gi", memoryLimit: "5Gi"},
			"suse-observability-hbase-hdfs-nn":        {cpuRequest: "2000m", cpuLimit: "4000m", memoryRequest: "1024Mi", memoryLimit: "2048Mi"},
			"suse-observability-hbase-hdfs-snn":       {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-tephra":         {cpuRequest: "4000m", cpuLimit: "6000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-kafka":                {cpuRequest: "4000m", cpuLimit: "5000m", memoryRequest: "6Gi", memoryLimit: "8Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "7", cpuLimit: "8000m", memoryRequest: "18Gi", memoryLimit: "18Gi"},
			"suse-observability-victoria-metrics-1":   {cpuRequest: "7", cpuLimit: "8000m", memoryRequest: "18Gi", memoryLimit: "18Gi"},
			"suse-observability-vmagent":              {cpuRequest: "4000m", cpuLimit: "5000m", memoryRequest: "5000Mi", memoryLimit: "5000Mi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "1000m", cpuLimit: "1500m", memoryRequest: "768Mi", memoryLimit: "768Mi"},
		},
		"default-nonha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-stackgraph":     {cpuRequest: "1000m", cpuLimit: "", memoryRequest: "2Gi", memoryLimit: "3Gi"},
			"suse-observability-hbase-tephra-mono":    {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-kafka":                {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "2Gi", memoryLimit: "2Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "300m", cpuLimit: "1", memoryRequest: "3584Mi", memoryLimit: "4Gi"},
			"suse-observability-vmagent":              {cpuRequest: "200m", cpuLimit: "200m", memoryRequest: "256Mi", memoryLimit: "512Mi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "100m", cpuLimit: "250m", memoryRequest: "640Mi", memoryLimit: "640Mi"},
		},
		"default-ha": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "1000m", cpuLimit: "2000m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hbase-master":   {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-hbase-rs":       {cpuRequest: "500m", cpuLimit: "3000m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-hbase-hdfs-dn":        {cpuRequest: "100m", cpuLimit: "500m", memoryRequest: "4Gi", memoryLimit: "4Gi"},
			"suse-observability-hbase-hdfs-nn":        {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-hdfs-snn":       {cpuRequest: "50m", cpuLimit: "500m", memoryRequest: "1Gi", memoryLimit: "1Gi"},
			"suse-observability-hbase-tephra":         {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "3Gi", memoryLimit: "3Gi"},
			"suse-observability-kafka":                {cpuRequest: "500m", cpuLimit: "1000m", memoryRequest: "2Gi", memoryLimit: "2Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "250m", cpuLimit: "500m", memoryRequest: "512Mi", memoryLimit: "512Mi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "300m", cpuLimit: "1", memoryRequest: "3584Mi", memoryLimit: "4Gi"},
			"suse-observability-victoria-metrics-1":   {cpuRequest: "300m", cpuLimit: "1", memoryRequest: "3584Mi", memoryLimit: "4Gi"},
			"suse-observability-vmagent":              {cpuRequest: "200m", cpuLimit: "200m", memoryRequest: "256Mi", memoryLimit: "512Mi"},
			"suse-observability-workload-observer":    {cpuRequest: "20m", cpuLimit: "50m", memoryRequest: "24Mi", memoryLimit: "128Mi"},
			"suse-observability-zookeeper":            {cpuRequest: "100m", cpuLimit: "250m", memoryRequest: "640Mi", memoryLimit: "640Mi"},
		},
		"10-nonha-overrides": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-hbase-stackgraph":     {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-hbase-tephra-mono":    {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-kafka":                {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-vmagent":              {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-workload-observer":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-zookeeper":            {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
		},
		"150-ha-overrides": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-hbase-hbase-master":   {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-hbase-hbase-rs":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-hbase-hdfs-dn":        {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-hbase-hdfs-nn":        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-hbase-hdfs-snn":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-hbase-tephra":         {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-kafka":                {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-victoria-metrics-1":   {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-vmagent":              {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-workload-observer":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-zookeeper":            {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
		},
		"4000-ha-overrides": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-hbase-hbase-master":   {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-hbase-hbase-rs":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-hbase-hdfs-dn":        {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-hbase-hdfs-nn":        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-hbase-hdfs-snn":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-hbase-tephra":         {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-kafka":                {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-victoria-metrics-1":   {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-vmagent":              {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-workload-observer":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-zookeeper":            {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
		},
		"10-nonha-global-overrides": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-hbase-stackgraph":     {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-hbase-tephra-mono":    {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-kafka":                {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-vmagent":              {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-workload-observer":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-zookeeper":            {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
		},
		"150-ha-global-overrides": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-hbase-hbase-master":   {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-hbase-hbase-rs":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-hbase-hdfs-dn":        {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-hbase-hdfs-nn":        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-hbase-hdfs-snn":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-hbase-tephra":         {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-kafka":                {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-victoria-metrics-1":   {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-vmagent":              {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-workload-observer":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-zookeeper":            {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
		},
		"4000-ha-global-overrides": {
			"suse-observability-clickhouse-shard0":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-elasticsearch-master": {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-hbase-hbase-master":   {cpuRequest: "222", cpuLimit: "222", memoryRequest: "222Gi", memoryLimit: "222Gi"},
			"suse-observability-hbase-hbase-rs":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-hbase-hdfs-dn":        {cpuRequest: "444", cpuLimit: "444", memoryRequest: "444Gi", memoryLimit: "444Gi"},
			"suse-observability-hbase-hdfs-nn":        {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-hbase-hdfs-snn":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-hbase-tephra":         {cpuRequest: "666", cpuLimit: "666", memoryRequest: "666Gi", memoryLimit: "666Gi"},
			"suse-observability-kafka":                {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
			"suse-observability-otel-collector":       {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-victoria-metrics-0":   {cpuRequest: "111", cpuLimit: "111", memoryRequest: "111Gi", memoryLimit: "111Gi"},
			"suse-observability-victoria-metrics-1":   {cpuRequest: "333", cpuLimit: "333", memoryRequest: "333Gi", memoryLimit: "333Gi"},
			"suse-observability-vmagent":              {cpuRequest: "555", cpuLimit: "555", memoryRequest: "555Gi", memoryLimit: "555Gi"},
			"suse-observability-workload-observer":    {cpuRequest: "999", cpuLimit: "999", memoryRequest: "999Gi", memoryLimit: "999Gi"},
			"suse-observability-zookeeper":            {cpuRequest: "777", cpuLimit: "777", memoryRequest: "777Gi", memoryLimit: "777Gi"},
		},
	}

	testCases := []struct {
		name         string
		valuesFiles  []string
		deployments  map[string]expectedResources
		statefulsets map[string]expectedResources
	}{
		{
			name:         "default-nonha",
			valuesFiles:  []string{"values/default_nonha_sizing.yaml"},
			deployments:  expectedDeployments["default-nonha"],
			statefulsets: expectedStatefulsets["default-nonha"],
		},
		{
			name:         "default-ha",
			valuesFiles:  []string{"values/default_ha_sizing.yaml"},
			deployments:  expectedDeployments["default-ha"],
			statefulsets: expectedStatefulsets["default-ha"],
		},
		{
			name:         "10-nonha",
			valuesFiles:  []string{"values/values_sizing_10_nonha.yaml"},
			deployments:  expectedDeployments["10-nonha"],
			statefulsets: expectedStatefulsets["10-nonha"],
		},
		{
			name:         "20-nonha",
			valuesFiles:  []string{"values/values_sizing_20_nonha.yaml"},
			deployments:  expectedDeployments["20-nonha"],
			statefulsets: expectedStatefulsets["20-nonha"],
		},
		{
			name:         "50-nonha",
			valuesFiles:  []string{"values/values_sizing_50_nonha.yaml"},
			deployments:  expectedDeployments["50-nonha"],
			statefulsets: expectedStatefulsets["50-nonha"],
		},
		{
			name:         "100-nonha",
			valuesFiles:  []string{"values/values_sizing_100_nonha.yaml"},
			deployments:  expectedDeployments["100-nonha"],
			statefulsets: expectedStatefulsets["100-nonha"],
		},
		{
			name:         "150-ha",
			valuesFiles:  []string{"values/values_sizing_150_ha.yaml"},
			deployments:  expectedDeployments["150-ha"],
			statefulsets: expectedStatefulsets["150-ha"],
		},
		{
			name:         "250-ha",
			valuesFiles:  []string{"values/values_sizing_250_ha.yaml"},
			deployments:  expectedDeployments["250-ha"],
			statefulsets: expectedStatefulsets["250-ha"],
		},
		{
			name:         "500-ha",
			valuesFiles:  []string{"values/values_sizing_500_ha.yaml"},
			deployments:  expectedDeployments["500-ha"],
			statefulsets: expectedStatefulsets["500-ha"],
		},
		{
			name:         "4000-ha",
			valuesFiles:  []string{"values/values_sizing_4000_ha.yaml"},
			deployments:  expectedDeployments["4000-ha"],
			statefulsets: expectedStatefulsets["4000-ha"],
		},
		{
			name:         "global-10-nonha",
			valuesFiles:  []string{"values/global_sizing_10_nonha.yaml"},
			deployments:  expectedDeployments["10-nonha"],
			statefulsets: expectedStatefulsets["10-nonha"],
		},
		{
			name:         "global-20-nonha",
			valuesFiles:  []string{"values/global_sizing_20_nonha.yaml"},
			deployments:  expectedDeployments["20-nonha"],
			statefulsets: expectedStatefulsets["20-nonha"],
		},
		{
			name:         "global-50-nonha",
			valuesFiles:  []string{"values/global_sizing_50_nonha.yaml"},
			deployments:  expectedDeployments["50-nonha"],
			statefulsets: expectedStatefulsets["50-nonha"],
		},
		{
			name:         "global-100-nonha",
			valuesFiles:  []string{"values/global_sizing_100_nonha.yaml"},
			deployments:  expectedDeployments["100-nonha"],
			statefulsets: expectedStatefulsets["100-nonha"],
		},
		{
			name:         "global-150-ha",
			valuesFiles:  []string{"values/global_sizing_150_ha.yaml"},
			deployments:  expectedDeployments["150-ha"],
			statefulsets: expectedStatefulsets["150-ha"],
		},
		{
			name:         "global-250-ha",
			valuesFiles:  []string{"values/global_sizing_250_ha.yaml"},
			deployments:  expectedDeployments["250-ha"],
			statefulsets: expectedStatefulsets["250-ha"],
		},
		{
			name:         "global-500-ha",
			valuesFiles:  []string{"values/global_sizing_500_ha.yaml"},
			deployments:  expectedDeployments["500-ha"],
			statefulsets: expectedStatefulsets["500-ha"],
		},
		{
			name:         "global-4000-ha",
			valuesFiles:  []string{"values/global_sizing_4000_ha.yaml"},
			deployments:  expectedDeployments["4000-ha"],
			statefulsets: expectedStatefulsets["4000-ha"],
		},
		{
			name:         "10-nonha-overrides",
			valuesFiles:  []string{"values/values_sizing_10_nonha.yaml", "values/resources_overrides.yaml"},
			deployments:  expectedDeployments["10-nonha-overrides"],
			statefulsets: expectedStatefulsets["10-nonha-overrides"],
		},
		{
			name:         "150-ha-overrides",
			valuesFiles:  []string{"values/values_sizing_150_ha.yaml", "values/resources_overrides.yaml"},
			deployments:  expectedDeployments["150-ha-overrides"],
			statefulsets: expectedStatefulsets["150-ha-overrides"],
		},
		{
			name:         "4000-ha-overrides",
			valuesFiles:  []string{"values/values_sizing_4000_ha.yaml", "values/resources_overrides.yaml"},
			deployments:  expectedDeployments["4000-ha-overrides"],
			statefulsets: expectedStatefulsets["4000-ha-overrides"],
		},
		{
			name:         "10-nonha-global-overrides",
			valuesFiles:  []string{"values/global_sizing_10_nonha.yaml", "values/resources_overrides.yaml"},
			deployments:  expectedDeployments["10-nonha-global-overrides"],
			statefulsets: expectedStatefulsets["10-nonha-global-overrides"],
		},
		{
			name:         "150-ha-global-overrides",
			valuesFiles:  []string{"values/global_sizing_150_ha.yaml", "values/resources_overrides.yaml"},
			deployments:  expectedDeployments["150-ha-global-overrides"],
			statefulsets: expectedStatefulsets["150-ha-global-overrides"],
		},
		{
			name:         "4000-ha-global-overrides",
			valuesFiles:  []string{"values/global_sizing_4000_ha.yaml", "values/resources_overrides.yaml"},
			deployments:  expectedDeployments["4000-ha-global-overrides"],
			statefulsets: expectedStatefulsets["4000-ha-global-overrides"],
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := helmtestutil.RenderHelmTemplate(t, "suse-observability", tc.valuesFiles...)
			resources := helmtestutil.NewKubernetesResources(t, output)

			// Check for extra/missing deployments
			t.Run("deployments-completeness", func(t *testing.T) {
				for name := range resources.Deployments {
					if _, ok := tc.deployments[name]; !ok {
						t.Errorf("Deployment %q is rendered by the chart but not covered by the test, please add it to the test case", name)
					}
				}
				for name := range tc.deployments {
					if _, ok := resources.Deployments[name]; !ok {
						t.Errorf("Deployment %q is listed in the test case but not rendered by the chart, please remove it from the test case", name)
					}
				}
			})

			// Check for extra/missing statefulsets
			t.Run("statefulsets-completeness", func(t *testing.T) {
				for name := range resources.Statefulsets {
					if _, ok := tc.statefulsets[name]; !ok {
						t.Errorf("StatefulSet %q is rendered by the chart but not covered by the test, please add it to the test case", name)
					}
				}
				for name := range tc.statefulsets {
					if _, ok := resources.Statefulsets[name]; !ok {
						t.Errorf("StatefulSet %q is listed in the test case but not rendered by the chart, please remove it from the test case", name)
					}
				}
			})

			for name, expected := range tc.deployments {
				t.Run("deployment-"+name, func(t *testing.T) {
					dep, exists := resources.Deployments[name]
					require.True(t, exists, "Deployment %s should exist", name)
					containers := dep.Spec.Template.Spec.Containers
					require.NotEmpty(t, containers)
					assertResources(t, "Deployment", name, containers[0].Resources, expected)
				})
			}

			for name, expected := range tc.statefulsets {
				t.Run("statefulset-"+name, func(t *testing.T) {
					ss, exists := resources.Statefulsets[name]
					require.True(t, exists, "StatefulSet %s should exist", name)
					containers := ss.Spec.Template.Spec.Containers
					require.NotEmpty(t, containers)
					assertResources(t, "StatefulSet", name, containers[0].Resources, expected)
				})
			}
		})
	}
}

func assertResources(t *testing.T, kind string, name string, actual corev1.ResourceRequirements, expected expectedResources) {
	t.Helper()
	if expected.memoryLimit != "" {
		memLimit := actual.Limits[corev1.ResourceMemory]
		assert.Equal(t, resource.MustParse(expected.memoryLimit), memLimit,
			"%s %s memory limit should be %s", kind, name, expected.memoryLimit)
	}
	if expected.cpuLimit != "" {
		cpuLimit := actual.Limits[corev1.ResourceCPU]
		assert.Equal(t, resource.MustParse(expected.cpuLimit), cpuLimit,
			"%s %s CPU limit should be %s", kind, name, expected.cpuLimit)
	}
	if expected.memoryRequest != "" {
		memRequest := actual.Requests[corev1.ResourceMemory]
		assert.Equal(t, resource.MustParse(expected.memoryRequest), memRequest,
			"%s %s memory request should be %s", kind, name, expected.memoryRequest)
	}
	if expected.cpuRequest != "" {
		cpuRequest := actual.Requests[corev1.ResourceCPU]
		assert.Equal(t, resource.MustParse(expected.cpuRequest), cpuRequest,
			"%s %s CPU request should be %s", kind, name, expected.cpuRequest)
	}
}
