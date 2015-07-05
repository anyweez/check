/**
 * This worker provides a parallelized sandbox for executing user-provided Javascript
 * in. Messages are posted from the core app that contain a task definition; the task
 * object is made fully available to the RankingFunc, which is executed with eval().
 * The outcome is reported back to the core app.
 */
self.addEventListener("message", function(e) {
	var task = JSON.parse(e.data);

	var response = {
		TaskId: task.TaskId,
		Score: 0.0,
		Valid: false,
	}
	// Run the ranking function and gather the response.
	try {
		func = eval("(function (task) {" + task.RankingFunc + "})")
		response.Score = Number( func(task) );

		if (response.Score != undefined) response.Valid = true;
	// Catch the error so execution continues but we don't need to do anything
	// with it since response.Valid is already set to false.
	} catch (e) {
		console.log(e);
	}

	self.postMessage(JSON.stringify(response));
})