function Prompt() {
    let toast = function(c) {
      const {
        msg = "",
        icon = "success",
        position = "top-end", 
      } = c;
      const Toast = Swal.mixin({
      toast: true,
      title: msg,
      icon: icon,
      position: position,
      showConfirmButton: false,
      timer: 3000,
      timerProgressBar: true,
      didOpen: (toast) => {
        toast.addEventListener('mouseenter', Swal.stopTimer)
        toast.addEventListener('mouseleave', Swal.resumeTimer)
      }
    })
      Toast.fire({})
    }
    let success = function(c) {
      const {
        title = "",
        msg = "You really should add something here",
        footer = "",
      } = c;
      Swal.fire({
        icon: 'success',
        title: title,
        text: msg,
        footer: footer
      })
    }
    let error = function(c) {
      const {
        title = "",
        msg = "You really should add something here",
        footer = "",
      } = c;
      Swal.fire({
        icon: 'error',
        title: title,
        text: msg,
        footer: footer
      })
    }
    async function custom(c) {
      const {
        icon = "",
        msg = "",
        title = "",
        showConfirmButton = true,
      } = c;
      const { value: result } = await Swal.fire({
        icon: icon,
        title: title,
        backdrop: false,
        html:msg,
        focusConfirm: false,
        showCancelButton: true,
        showConfirmButton: showConfirmButton,
        willOpen: () => {
          if(c.willOpen !== undefined) {
            c.willOpen();
          }
        },
      })

      if (result) {
        if (result.dismiss !== Swal.DismissReason.cancel) {
          if (result.value !== "") {
            if (c.callback !== undefined) {
              c.callback(result);
            } else {
              c.callback(false);
            }                    
          }
        } else {
          c.callback(false);
        }
      }
      
    }
    return {
      toast: toast,
      success: success,
      error: error,
      custom: custom,
    }
  }

//PopUp is used to display a modal dialog on room pages for searching availablity, must take csrf token and roomId 
function PopUp(token, roomId){
    document.getElementById("check-availability-btn").addEventListener("click", function () {
        let html = `
        <form id="check-availability-form" action="" method="post" novalidate>
        <div class="row">
            <div class="col">
                <div class="row" id="reservation-dates-modal">
                    <div class="col">
                        <input enabled required class="form-control" type="text" name="start" id="start" placeholder="Arrival">
                    </div>
                    <div class="col">
                        <input enabled required class="form-control" type="text" name="end" id="end" placeholder="Departure">
                    </div>
                </div>
            </div>
        </div>
        </form>
        `
        attention.custom({
            msg: html,
            title: "Choose your dates",

            willOpen: () => {
                const elem = document.getElementById('reservation-dates-modal');
                const rp = new DateRangePicker(elem, {
                    format: 'yyyy-mm-dd',
                    showOnFocus: false,
                    orientation: 'top',
                    minDate: new Date(),
                })
            },

            callback: function (result) {
                let form = document.getElementById("check-availability-form");
                let formData = new FormData(form);
                formData.append("csrf_token", token);
                formData.append("room_id", roomId.toString());

                fetch('/search_availability-json', {
                    method: "post",
                    body: formData,
                })
                    .then(response => response.json())
                    .then(data => {
                        if (data.ok) {
                            attention.custom({
                                icon: 'success',
                                msg: '<p>Room is available</p>'
                                + '<p><a href="/book_room?id=' + data.room_id + '&s=' + data.start_date + '&e=' + data.end_date +  
                                '"class="btn btn-primary">' + 'Book now</a></p>',
                                showConfirmButton: false
                            })
                        } else {
                            attention.error({
                                msg: data.message,
                            })
                        }
                    })
            }
        });
    })

}


function processRes(id, src) {
  attention.custom({
    icon: "warning",
    msg: "Are you sure?",
    callback: function(result) {
      if (result !== false) {
        window.location.href = "/admin/process_reservation/" + src + "/" + id
      }
    }
  })
}