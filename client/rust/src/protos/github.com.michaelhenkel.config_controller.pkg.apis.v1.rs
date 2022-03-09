#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SubscriptionRequest {
    #[prost(string, tag = "1")]
    pub name: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Response {
    #[prost(message, optional, tag = "1")]
    pub new: ::core::option::Option<Resource>,
    #[prost(message, optional, tag = "2")]
    pub old: ::core::option::Option<Resource>,
    #[prost(enumeration = "response::Action", tag = "3")]
    pub action: i32,
}
/// Nested message and enum types in `Response`.
pub mod response {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Action {
        Add = 0,
        Update = 1,
        Delete = 2,
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Resource {
    #[prost(oneof = "resource::Resource", tags = "1, 2, 3, 4, 5, 6, 7")]
    pub resource: ::core::option::Option<resource::Resource>,
}
/// Nested message and enum types in `Resource`.
pub mod resource {
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Resource {
        #[prost(message, tag="1")]
        VirtualNetwork(super::super::super::super::super::super::super::super::ssd_git::juniper::net::contrail::cn2::contrail::pkg::apis::core::v1alpha1::VirtualNetwork),
        #[prost(message, tag="2")]
        VirtualMachineInterface(super::super::super::super::super::super::super::super::ssd_git::juniper::net::contrail::cn2::contrail::pkg::apis::core::v1alpha1::VirtualMachineInterface),
        #[prost(message, tag="3")]
        VirtualRouter(super::super::super::super::super::super::super::super::ssd_git::juniper::net::contrail::cn2::contrail::pkg::apis::core::v1alpha1::VirtualRouter),
        #[prost(message, tag="4")]
        VirtualMachine(super::super::super::super::super::super::super::super::ssd_git::juniper::net::contrail::cn2::contrail::pkg::apis::core::v1alpha1::VirtualMachine),
        #[prost(message, tag="5")]
        InstanceIp(super::super::super::super::super::super::super::super::ssd_git::juniper::net::contrail::cn2::contrail::pkg::apis::core::v1alpha1::InstanceIp),
        #[prost(message, tag="6")]
        RoutingInstance(super::super::super::super::super::super::super::super::ssd_git::juniper::net::contrail::cn2::contrail::pkg::apis::core::v1alpha1::RoutingInstance),
        #[prost(message, tag="7")]
        Namespace(super::super::super::super::super::super::super::super::k8s::io::api::core::v1::Namespace),
    }
}
#[doc = r" Generated client implementations."]
pub mod config_controller_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    #[derive(Debug, Clone)]
    pub struct ConfigControllerClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl ConfigControllerClient<tonic::transport::Channel> {
        #[doc = r" Attempt to create a new client by connecting to a given endpoint."]
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: std::convert::TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> ConfigControllerClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::BoxBody>,
        T::ResponseBody: Body + Send + 'static,
        T::Error: Into<StdError>,
        <T::ResponseBody as Body>::Error: Into<StdError> + Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> ConfigControllerClient<InterceptedService<T, F>>
        where
            F: tonic::service::Interceptor,
            T: tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
                Response = http::Response<
                    <T as tonic::client::GrpcService<tonic::body::BoxBody>>::ResponseBody,
                >,
            >,
            <T as tonic::codegen::Service<http::Request<tonic::body::BoxBody>>>::Error:
                Into<StdError> + Send + Sync,
        {
            ConfigControllerClient::new(InterceptedService::new(inner, interceptor))
        }
        #[doc = r" Compress requests with `gzip`."]
        #[doc = r""]
        #[doc = r" This requires the server to support it otherwise it might respond with an"]
        #[doc = r" error."]
        pub fn send_gzip(mut self) -> Self {
            self.inner = self.inner.send_gzip();
            self
        }
        #[doc = r" Enable decompressing responses with `gzip`."]
        pub fn accept_gzip(mut self) -> Self {
            self.inner = self.inner.accept_gzip();
            self
        }
        pub async fn subscribe_list_watch(
            &mut self,
            request: impl tonic::IntoRequest<super::SubscriptionRequest>,
        ) -> Result<tonic::Response<tonic::codec::Streaming<super::Response>>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http :: uri :: PathAndQuery :: from_static ("/github.com.michaelhenkel.config_controller.pkg.apis.v1.ConfigController/SubscribeListWatch") ;
            self.inner
                .server_streaming(request.into_request(), path, codec)
                .await
        }
    }
}
#[doc = r" Generated server implementations."]
pub mod config_controller_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    #[doc = "Generated trait containing gRPC methods that should be implemented for use with ConfigControllerServer."]
    #[async_trait]
    pub trait ConfigController: Send + Sync + 'static {
        #[doc = "Server streaming response type for the SubscribeListWatch method."]
        type SubscribeListWatchStream: futures_core::Stream<Item = Result<super::Response, tonic::Status>>
            + Send
            + 'static;
        async fn subscribe_list_watch(
            &self,
            request: tonic::Request<super::SubscriptionRequest>,
        ) -> Result<tonic::Response<Self::SubscribeListWatchStream>, tonic::Status>;
    }
    #[derive(Debug)]
    pub struct ConfigControllerServer<T: ConfigController> {
        inner: _Inner<T>,
        accept_compression_encodings: (),
        send_compression_encodings: (),
    }
    struct _Inner<T>(Arc<T>);
    impl<T: ConfigController> ConfigControllerServer<T> {
        pub fn new(inner: T) -> Self {
            let inner = Arc::new(inner);
            let inner = _Inner(inner);
            Self {
                inner,
                accept_compression_encodings: Default::default(),
                send_compression_encodings: Default::default(),
            }
        }
        pub fn with_interceptor<F>(inner: T, interceptor: F) -> InterceptedService<Self, F>
        where
            F: tonic::service::Interceptor,
        {
            InterceptedService::new(Self::new(inner), interceptor)
        }
    }
    impl<T, B> tonic::codegen::Service<http::Request<B>> for ConfigControllerServer<T>
    where
        T: ConfigController,
        B: Body + Send + 'static,
        B::Error: Into<StdError> + Send + 'static,
    {
        type Response = http::Response<tonic::body::BoxBody>;
        type Error = Never;
        type Future = BoxFuture<Self::Response, Self::Error>;
        fn poll_ready(&mut self, _cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
            Poll::Ready(Ok(()))
        }
        fn call(&mut self, req: http::Request<B>) -> Self::Future {
            let inner = self.inner.clone();
            match req . uri () . path () { "/github.com.michaelhenkel.config_controller.pkg.apis.v1.ConfigController/SubscribeListWatch" => { # [allow (non_camel_case_types)] struct SubscribeListWatchSvc < T : ConfigController > (pub Arc < T >) ; impl < T : ConfigController > tonic :: server :: ServerStreamingService < super :: SubscriptionRequest > for SubscribeListWatchSvc < T > { type Response = super :: Response ; type ResponseStream = T :: SubscribeListWatchStream ; type Future = BoxFuture < tonic :: Response < Self :: ResponseStream > , tonic :: Status > ; fn call (& mut self , request : tonic :: Request < super :: SubscriptionRequest >) -> Self :: Future { let inner = self . 0 . clone () ; let fut = async move { (* inner) . subscribe_list_watch (request) . await } ; Box :: pin (fut) } } let accept_compression_encodings = self . accept_compression_encodings ; let send_compression_encodings = self . send_compression_encodings ; let inner = self . inner . clone () ; let fut = async move { let inner = inner . 0 ; let method = SubscribeListWatchSvc (inner) ; let codec = tonic :: codec :: ProstCodec :: default () ; let mut grpc = tonic :: server :: Grpc :: new (codec) . apply_compression_config (accept_compression_encodings , send_compression_encodings) ; let res = grpc . server_streaming (method , req) . await ; Ok (res) } ; Box :: pin (fut) } _ => Box :: pin (async move { Ok (http :: Response :: builder () . status (200) . header ("grpc-status" , "12") . header ("content-type" , "application/grpc") . body (empty_body ()) . unwrap ()) }) , }
        }
    }
    impl<T: ConfigController> Clone for ConfigControllerServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self {
                inner,
                accept_compression_encodings: self.accept_compression_encodings,
                send_compression_encodings: self.send_compression_encodings,
            }
        }
    }
    impl<T: ConfigController> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(self.0.clone())
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: ConfigController> tonic::transport::NamedService for ConfigControllerServer<T> {
        const NAME: &'static str =
            "github.com.michaelhenkel.config_controller.pkg.apis.v1.ConfigController";
    }
}
