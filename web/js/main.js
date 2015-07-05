var vm;

$(document).ready(function() {
	vm = new ViewModel();
	ko.applyBindings(vm);

	var worker = new Worker("static/js/worker.js");
	// Fetch the initial list of tasks.
	FetchTasks()

	$("#filter-save").on("click", function() {
		FetchTasks()
	});
	// Save button for ranking function. This should
	$("#ranking-save").on("click", function() {
		for (var taskId in vm.Tasks()) {
			var task = vm.Tasks()[taskId];
			task.RankingFunc = vm.RankingFunc();

			worker.postMessage(JSON.stringify(task));			
		}
	})

	worker.addEventListener("message", UpdateScore, false);
});

function FetchTasks() {
	console.log("Fetching tasks");
	// Fetch data
	$.ajax({
		url: "/app/fetch?q=" + encodeURIComponent(vm.FilterString()),
		success: HandleTasks,
		// TODO: need to add a failure case.
	});	
}


/**
 * When we first receive a list of tasks this is what we get working on.
 * It provides the overall initialization flow after data has been received.
 */
function HandleTasks(taskList) {
	// Clear the existing tasks out of the array.
	vm.Tasks().length = 0;

	// Add some additional fields to the tasks objects and them insert them
	// into the observableArray.
	for (var i = 0; i < taskList.length; i++) {
		taskList[i].TaskId = hex_sha256(taskList[i].Title + taskList[i].Snippet);
		taskList[i].Score = Score = ko.observable(0);

		taskList.Labels = ko.computed(function() {
			return "Labels: " + this.Labels.join(", ");
		}, taskList[i]);
		// Flag identifying whether this tag has been marked as complete. Certain
		// logic, such as marking off items in the UI, will be based on it.
		taskList[i].Completed = false;

		vm.Tasks.push(taskList[i]);
	}
}

function ViewModel(taskList) {
	// Tasks aren't populated until the AJAX request completes (see HandleTasks).
	this.Tasks = ko.observableArray();
	// Filter string is passed w/ AJAX requests to limit what's defined as a "task."
	this.FilterString = ko.observable("");
	// RankingFunc is executed in a sandbox on the client.
	this.RankingFunc = ko.observable("return task.Labels.length + 5;");
	// A flag indicating whether the text in the ranking function is executable.
	this.RankingFuncValid = true;
}

/**
 * Receives a task, produces a score.
 */
function UserFunction(task) {
	return task.Labels.length;
}

/**
 * Function that returns the priority score for a given task.
 */
function ComputeScore(task) {
	return 15.3;
}


// Recompute all scores when a new function is provided or paramters
// change.
function UpdateScore(e) {
	var response = JSON.parse(e.data);

	// Mark the ranking function as invalid if the worker couldn't execute
	// it successfully.
	vm.RankingFuncValid = response.Valid;
	if (!response.Valid) {
		console.log("Invalid ranking function.");
		return;
	}

	for (var i in vm.Tasks()) {
		if (vm.Tasks()[i].TaskId == response.TaskId) {
			vm.Tasks()[i].Score(response.Score);
		}
	}

	// Resort.
	vm.Tasks.sort(function(left, right) {
		return right.Score() - left.Score();
	})
}