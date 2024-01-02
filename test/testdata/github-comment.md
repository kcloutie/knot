<!-- =================================== The following sets the values of variables used throughout the template =================================== -->
{{ $pipelineStatus := "Failed :x:" -}}
{{if eq .data.extra.pipelineRunVariables.TN_PIPELINE_STATUS "Successful" -}}
  {{$pipelineStatus = "Succeeded :white_check_mark:" -}}
{{end -}}

{{ $tasksToLogNames := "" | splitList "," -}}
{{ $tasksToLog := newArray -}}
{{ $planSumRegex := "Plan: [0-9]+ to add, [0-9]+ to change, [0-9]+ to destroy" -}}
{{ $planTask := "" -}}
{{ $planTaskName := .data.extra.pipelineRunObject.metadata.annotations | select "tekton-notify.internal.com/git-plan-task-name" | firstOrDefault "" -}}
{{ $commentTaskNames := .data.extra.pipelineRunObject.metadata.annotations | select "tekton-notify.internal.com/git-comment-task-names" | firstOrDefault "{}" | fromJson -}}

{{ range $name := $commentTaskNames -}}
   {{ $tasksToLogNames = append $tasksToLogNames $name -}}
{{ end -}}

{{ $failedTasks := .data.extra.pipelineRunVariables.TN_FAILED_TASKS | splitList "," -}}
{{ range $name := $failedTasks -}}
   {{ $tasksToLogNames = append $tasksToLogNames $name -}}
{{ end -}}

{{ $canceledTasks := .data.extra.pipelineRunVariables.TN_CANCELLED_TASKS | splitList "," -}}
{{ range $name := $canceledTasks -}}
   {{ $tasksToLogNames = append $tasksToLogNames $name -}}
{{ end -}}

{{ range $taskRun := .data.extra.taskRunDetails -}}
  {{ if has $taskRun.taskName $tasksToLogNames -}}
    {{ $tasksToLog = append $tasksToLog $taskRun -}}
  {{ end -}}
  {{ if eq $taskRun.taskName $planTaskName -}}
    {{ $planTask = $taskRun -}}
  {{ end -}}
{{ end -}}
<!-- =================================== Start of markdown content =================================== -->
{{ .data.extra.pipelineRunVariables.RESULT_COMMENT_HEADING }}

| Namespace | PipelineRun Name | Status | Runtime |  Log File |
| :- | :-: | :-: | :-: | :-: |
| {{ .data.namespace }} | {{ .data.pipelineRunName }} | {{ $pipelineStatus }} | {{ .data.pipelineRuntime }} |[Log File]({{ .data.pipelineRunLogFileUrl }}) |

{{if $planTask -}}
  {{ if (regexMatch $planSumRegex $planTask.logs) -}}
    {{ $summary := regexFind $planSumRegex $planTask.logs -}}
    {{if $summary -}}
      {{ $summary = $summary | replace "Plan: " "Plan: :white_check_mark: " -}}
      {{ $summary = $summary | replace "add, " "add, :large_orange_diamond: " -}}
      {{ $summary = $summary | replace "change, " "change, :boom: " -}}

{{ $summary -}}{{ printf "\n\n" -}}

    {{end -}}
  {{end -}}
{{end -}}

{{ range $taskRun := $tasksToLog -}}
  {{ $taskStatus := "Failed :x:" -}}
  {{if or (eq $taskRun.reason "Successful") (eq $taskRun.reason "Succeeded") -}}
    {{$taskStatus = "Succeeded :white_check_mark:" -}}
  {{end -}}

<details><summary>Expand for <b>{{ $taskRun.taskName }}</b> Results {{ $taskStatus }}</summary>
<p>

| Name | Status | Start | Completed | Total Time |
| :- | :-: | :-: | :-: | :-: |
| {{ $taskRun.taskName }} | {{ $taskStatus }} | {{ $taskRun.startedOn }} | {{ $taskRun.completedOn }} | {{ $taskRun.totalTime }} |

**Tasks details below may be truncated.  If so, refer to full log above.**

```powershell
{{ $taskRun.logs }}
```

</p>
</details>

{{ end -}}
