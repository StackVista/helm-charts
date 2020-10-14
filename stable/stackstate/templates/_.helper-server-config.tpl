{{- define "stackstate.server.memory.resource" -}}
{{- $baseMemoryConsumption := "380Mi" -}}
{{- $heapMemoryFraction := 75 -}}
{{- $memoryLimitsMB := (include "stackstate.storage.to.megabytes" .) -}}
{{- $baseMemoryConsumptionMB := (include "stackstate.storage.to.megabytes" $baseMemoryConsumption) -}}
{{- $heapMemory := (sub $memoryLimitsMB  $baseMemoryConsumptionMB) | int -}}
{{- (div (mul $heapMemory $heapMemoryFraction) 100) | int -}}
{{- end -}}
