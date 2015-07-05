$(document).ready(function() {
	console.log(document.cookie);
	// Fetch data
	var data = $.ajax({
		url: "/app/fetch",
		success: HandleTasks,
		// TODO: need to add a failure case.
	});
});

/**
 * When we first receive a list of tasks this is what we get working on.
 * It provides the overall initialization flow after data has been received.
 */
function HandleTasks(taskList) {
	console.log(taskList);

	ko.applyBindings(new CreateViewModel(taskList));
}

function CreateViewModel(taskList) {
	this.tasks = taskList;

	for (var i = 0; i < taskList.length; i++) {
		this.tasks[i].Score = ko.computed(function () {
			return UserFunction(this);
		}, this.tasks[i]);

		this.tasks[i].Labels = ko.computed(function() {
			return "Labels: " + this.Labels.join(", ");
		}, this.tasks[i]);
		// Flag identifying whether this tag has been marked as complete. Certain
		// logic, such as marking off items in the UI, will be based on it.
		this.tasks[i].Completed = false;
	}
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
	console.log(this);
	return 15.3;
}


// Recompute all scores when a new function is provided or paramters
// change.
function UpdateScores() {

}