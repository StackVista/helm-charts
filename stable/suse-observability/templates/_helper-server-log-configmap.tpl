{{/*
Shared settings in configmap for logging on stackstate sync pods
*/}}
{{- define "stackstate.configmap.server-base-log" }}
    <logger name="com.stackstate" level="INFO"/>
    <logger name="org.apache.kafka.common.utils.AppInfoParser" level="ERROR"/>
    <logger name="com.stackvista.graph" level="WARN"/>
    <logger name="co.cask.tephra" level="WARN"/>
    <logger name="com.stackvista.graph.migration" level="INFO"/>
    <logger name="com.stackvista.graph.core.slicing" level="INFO"/>
    <logger name="com.gilt.handlebars.scala" level="ERROR"/>
    <logger name="com.stackstate.sync" level="INFO"/>
    <logger name="com.stackstate.stackpackmanager" level="INFO"/>
    <logger name="com.stackstate.domain.stackelement.check.Check" level="DEBUG"/>
    <logger name="com.stackstate.domain.view.eventhandlers.EventHandler" level="DEBUG"/>
    <logger name="com.stackstate.reactivefunction.viewstateconfiguration.ViewHealthStateUpdaterDefinition" level="INFO"/>
    <logger name="com.stackstate.domain.sync.MappingFunction" level="DEBUG"/>
    <logger name="com.stackstate.domain.template.TemplateFunction" level="DEBUG"/>
    <logger name="com.stackstate.util.logging" level="INFO"/>
    <logger name="com.stackstate.util.logging.Check" level="INFO"/>
    <logger name="com.stackvista.graph.transaction.StackTransactionManager" level="INFO"/>
    <logger name="com.stackstate.sync.SyncLogging" level="INFO"/>
    <logger name="com.stackstate.sync.ExtTopoLogging" level="INFO"/>
    <logger name="com.stackstate.stackpackmanager.StackPackLogging" level="INFO"/>
    <logger name="com.stackstate.api.streamapi.vieweventsummary.ViewAccessLogging" level="INFO"/>
{{- end -}}
