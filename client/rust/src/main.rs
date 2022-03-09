use config_client::protos::github::com::michaelhenkel::config_controller::pkg::apis::v1::config_controller_client::ConfigControllerClient;
use config_client::protos::github::com::michaelhenkel::config_controller::pkg::apis::v1::SubscriptionRequest;
use config_client::protos::github::com::michaelhenkel::config_controller::pkg::apis::v1;
use tonic::transport::Channel;
use std::error::Error;
use std::env;
mod resources;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut client = ConfigControllerClient::connect("http://127.0.0.1:20443").await.unwrap();
    consume_response(&mut client).await?;

    println!("Connected...now sleeping for 2 seconds...");
    tokio::time::sleep(std::time::Duration::from_secs(2)).await;
    drop(client);

    println!("Disconnected...");
    Ok(())
}

fn get_node() -> String {
    if env::args().len() > 0 {
        let args: Vec<String> = env::args().collect();
        args[1].to_string()
    } else {
        "5b3s30".to_string()
    }
}

async fn consume_response(client: &mut ConfigControllerClient<Channel>) -> Result<(), Box<dyn Error>> {
    let request = tonic::Request::new(SubscriptionRequest {
        name: get_node(),
    });
    let mut stream = client
        .subscribe_list_watch(request)
        .await?
        .into_inner();
    while let Some(response) = stream.message().await? {
        let action = response.action();
        let res = resources::get_resource(response);
        match action {
            v1::response::Action::Add => res.add(),
            v1::response::Action::Update => res.update(),
            v1::response::Action::Delete => res.delete(),
        }        
    }
    drop(stream);
    Ok(())
}