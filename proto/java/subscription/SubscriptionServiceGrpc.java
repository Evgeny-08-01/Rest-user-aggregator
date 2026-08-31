package subscription;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * ============================================================
 * 5. СЕРВИС (главная часть — список всех RPC-методов)
 * ============================================================
 * Service — это интерфейс. Он описывает ВСЕ методы API, которые будут доступны
 * через gRPC.
 * Каждый rpc — это один метод.
 * Формат: rpc &lt;ИмяМетода&gt; (Запрос) returns (Ответ);
 * Имена методов должны быть глаголами (Create, Get, List, Update, Delete).
 * Это общепринятое соглашение для gRPC.
 * !!! ВАЖНО: методы делятся на 2 группы по смыслу:
 *   1. Работа с ПОДПИСКАМИ (CRUD) — для пользователей
 *   2. Работа с ШАБЛОНАМИ (CRUD) — для администраторов
 * Разделение такое же, как в REST-эндпоинтах:
 *   GET/POST /api/subscriptions      → работа с подписками
 *   GET/POST /api/admin/templates    → работа с шаблонами (только админ)
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.1)",
    comments = "Source: subscription.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class SubscriptionServiceGrpc {

  private SubscriptionServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "subscription.SubscriptionService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetSubscriptionsRequest,
      subscription.SubscriptionOuterClass.SubscriptionList> getGetSubscriptionsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetSubscriptions",
      requestType = subscription.SubscriptionOuterClass.GetSubscriptionsRequest.class,
      responseType = subscription.SubscriptionOuterClass.SubscriptionList.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetSubscriptionsRequest,
      subscription.SubscriptionOuterClass.SubscriptionList> getGetSubscriptionsMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetSubscriptionsRequest, subscription.SubscriptionOuterClass.SubscriptionList> getGetSubscriptionsMethod;
    if ((getGetSubscriptionsMethod = SubscriptionServiceGrpc.getGetSubscriptionsMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getGetSubscriptionsMethod = SubscriptionServiceGrpc.getGetSubscriptionsMethod) == null) {
          SubscriptionServiceGrpc.getGetSubscriptionsMethod = getGetSubscriptionsMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.GetSubscriptionsRequest, subscription.SubscriptionOuterClass.SubscriptionList>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetSubscriptions"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.GetSubscriptionsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.SubscriptionList.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("GetSubscriptions"))
              .build();
        }
      }
    }
    return getGetSubscriptionsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.CreateRequest,
      subscription.SubscriptionOuterClass.CreateResponse> getCreateSubscriptionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateSubscription",
      requestType = subscription.SubscriptionOuterClass.CreateRequest.class,
      responseType = subscription.SubscriptionOuterClass.CreateResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.CreateRequest,
      subscription.SubscriptionOuterClass.CreateResponse> getCreateSubscriptionMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.CreateRequest, subscription.SubscriptionOuterClass.CreateResponse> getCreateSubscriptionMethod;
    if ((getCreateSubscriptionMethod = SubscriptionServiceGrpc.getCreateSubscriptionMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getCreateSubscriptionMethod = SubscriptionServiceGrpc.getCreateSubscriptionMethod) == null) {
          SubscriptionServiceGrpc.getCreateSubscriptionMethod = getCreateSubscriptionMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.CreateRequest, subscription.SubscriptionOuterClass.CreateResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateSubscription"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.CreateRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.CreateResponse.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("CreateSubscription"))
              .build();
        }
      }
    }
    return getCreateSubscriptionMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest,
      subscription.SubscriptionOuterClass.Subscription> getGetSubscriptionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetSubscription",
      requestType = subscription.SubscriptionOuterClass.GetRequest.class,
      responseType = subscription.SubscriptionOuterClass.Subscription.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest,
      subscription.SubscriptionOuterClass.Subscription> getGetSubscriptionMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest, subscription.SubscriptionOuterClass.Subscription> getGetSubscriptionMethod;
    if ((getGetSubscriptionMethod = SubscriptionServiceGrpc.getGetSubscriptionMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getGetSubscriptionMethod = SubscriptionServiceGrpc.getGetSubscriptionMethod) == null) {
          SubscriptionServiceGrpc.getGetSubscriptionMethod = getGetSubscriptionMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.GetRequest, subscription.SubscriptionOuterClass.Subscription>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetSubscription"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.GetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.Subscription.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("GetSubscription"))
              .build();
        }
      }
    }
    return getGetSubscriptionMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.UpdateRequest,
      subscription.SubscriptionOuterClass.Empty> getUpdateSubscriptionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateSubscription",
      requestType = subscription.SubscriptionOuterClass.UpdateRequest.class,
      responseType = subscription.SubscriptionOuterClass.Empty.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.UpdateRequest,
      subscription.SubscriptionOuterClass.Empty> getUpdateSubscriptionMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.UpdateRequest, subscription.SubscriptionOuterClass.Empty> getUpdateSubscriptionMethod;
    if ((getUpdateSubscriptionMethod = SubscriptionServiceGrpc.getUpdateSubscriptionMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getUpdateSubscriptionMethod = SubscriptionServiceGrpc.getUpdateSubscriptionMethod) == null) {
          SubscriptionServiceGrpc.getUpdateSubscriptionMethod = getUpdateSubscriptionMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.UpdateRequest, subscription.SubscriptionOuterClass.Empty>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateSubscription"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.UpdateRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.Empty.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("UpdateSubscription"))
              .build();
        }
      }
    }
    return getUpdateSubscriptionMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest,
      subscription.SubscriptionOuterClass.Empty> getDeleteSubscriptionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DeleteSubscription",
      requestType = subscription.SubscriptionOuterClass.GetRequest.class,
      responseType = subscription.SubscriptionOuterClass.Empty.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest,
      subscription.SubscriptionOuterClass.Empty> getDeleteSubscriptionMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest, subscription.SubscriptionOuterClass.Empty> getDeleteSubscriptionMethod;
    if ((getDeleteSubscriptionMethod = SubscriptionServiceGrpc.getDeleteSubscriptionMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getDeleteSubscriptionMethod = SubscriptionServiceGrpc.getDeleteSubscriptionMethod) == null) {
          SubscriptionServiceGrpc.getDeleteSubscriptionMethod = getDeleteSubscriptionMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.GetRequest, subscription.SubscriptionOuterClass.Empty>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DeleteSubscription"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.GetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.Empty.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("DeleteSubscription"))
              .build();
        }
      }
    }
    return getDeleteSubscriptionMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.TotalCostRequest,
      subscription.SubscriptionOuterClass.TotalCostResponse> getGetTotalCostMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetTotalCost",
      requestType = subscription.SubscriptionOuterClass.TotalCostRequest.class,
      responseType = subscription.SubscriptionOuterClass.TotalCostResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.TotalCostRequest,
      subscription.SubscriptionOuterClass.TotalCostResponse> getGetTotalCostMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.TotalCostRequest, subscription.SubscriptionOuterClass.TotalCostResponse> getGetTotalCostMethod;
    if ((getGetTotalCostMethod = SubscriptionServiceGrpc.getGetTotalCostMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getGetTotalCostMethod = SubscriptionServiceGrpc.getGetTotalCostMethod) == null) {
          SubscriptionServiceGrpc.getGetTotalCostMethod = getGetTotalCostMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.TotalCostRequest, subscription.SubscriptionOuterClass.TotalCostResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetTotalCost"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.TotalCostRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.TotalCostResponse.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("GetTotalCost"))
              .build();
        }
      }
    }
    return getGetTotalCostMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.Empty,
      subscription.SubscriptionOuterClass.TemplateList> getListTemplatesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListTemplates",
      requestType = subscription.SubscriptionOuterClass.Empty.class,
      responseType = subscription.SubscriptionOuterClass.TemplateList.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.Empty,
      subscription.SubscriptionOuterClass.TemplateList> getListTemplatesMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.Empty, subscription.SubscriptionOuterClass.TemplateList> getListTemplatesMethod;
    if ((getListTemplatesMethod = SubscriptionServiceGrpc.getListTemplatesMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getListTemplatesMethod = SubscriptionServiceGrpc.getListTemplatesMethod) == null) {
          SubscriptionServiceGrpc.getListTemplatesMethod = getListTemplatesMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.Empty, subscription.SubscriptionOuterClass.TemplateList>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListTemplates"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.Empty.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.TemplateList.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("ListTemplates"))
              .build();
        }
      }
    }
    return getListTemplatesMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest,
      subscription.SubscriptionOuterClass.Template> getGetTemplateMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetTemplate",
      requestType = subscription.SubscriptionOuterClass.GetRequest.class,
      responseType = subscription.SubscriptionOuterClass.Template.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest,
      subscription.SubscriptionOuterClass.Template> getGetTemplateMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest, subscription.SubscriptionOuterClass.Template> getGetTemplateMethod;
    if ((getGetTemplateMethod = SubscriptionServiceGrpc.getGetTemplateMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getGetTemplateMethod = SubscriptionServiceGrpc.getGetTemplateMethod) == null) {
          SubscriptionServiceGrpc.getGetTemplateMethod = getGetTemplateMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.GetRequest, subscription.SubscriptionOuterClass.Template>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetTemplate"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.GetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.Template.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("GetTemplate"))
              .build();
        }
      }
    }
    return getGetTemplateMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.CreateTemplateRequest,
      subscription.SubscriptionOuterClass.CreateResponse> getCreateTemplateMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateTemplate",
      requestType = subscription.SubscriptionOuterClass.CreateTemplateRequest.class,
      responseType = subscription.SubscriptionOuterClass.CreateResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.CreateTemplateRequest,
      subscription.SubscriptionOuterClass.CreateResponse> getCreateTemplateMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.CreateTemplateRequest, subscription.SubscriptionOuterClass.CreateResponse> getCreateTemplateMethod;
    if ((getCreateTemplateMethod = SubscriptionServiceGrpc.getCreateTemplateMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getCreateTemplateMethod = SubscriptionServiceGrpc.getCreateTemplateMethod) == null) {
          SubscriptionServiceGrpc.getCreateTemplateMethod = getCreateTemplateMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.CreateTemplateRequest, subscription.SubscriptionOuterClass.CreateResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateTemplate"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.CreateTemplateRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.CreateResponse.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("CreateTemplate"))
              .build();
        }
      }
    }
    return getCreateTemplateMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.UpdateTemplateRequest,
      subscription.SubscriptionOuterClass.Empty> getUpdateTemplateMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateTemplate",
      requestType = subscription.SubscriptionOuterClass.UpdateTemplateRequest.class,
      responseType = subscription.SubscriptionOuterClass.Empty.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.UpdateTemplateRequest,
      subscription.SubscriptionOuterClass.Empty> getUpdateTemplateMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.UpdateTemplateRequest, subscription.SubscriptionOuterClass.Empty> getUpdateTemplateMethod;
    if ((getUpdateTemplateMethod = SubscriptionServiceGrpc.getUpdateTemplateMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getUpdateTemplateMethod = SubscriptionServiceGrpc.getUpdateTemplateMethod) == null) {
          SubscriptionServiceGrpc.getUpdateTemplateMethod = getUpdateTemplateMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.UpdateTemplateRequest, subscription.SubscriptionOuterClass.Empty>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateTemplate"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.UpdateTemplateRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.Empty.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("UpdateTemplate"))
              .build();
        }
      }
    }
    return getUpdateTemplateMethod;
  }

  private static volatile io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest,
      subscription.SubscriptionOuterClass.Empty> getDeleteTemplateMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DeleteTemplate",
      requestType = subscription.SubscriptionOuterClass.GetRequest.class,
      responseType = subscription.SubscriptionOuterClass.Empty.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest,
      subscription.SubscriptionOuterClass.Empty> getDeleteTemplateMethod() {
    io.grpc.MethodDescriptor<subscription.SubscriptionOuterClass.GetRequest, subscription.SubscriptionOuterClass.Empty> getDeleteTemplateMethod;
    if ((getDeleteTemplateMethod = SubscriptionServiceGrpc.getDeleteTemplateMethod) == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        if ((getDeleteTemplateMethod = SubscriptionServiceGrpc.getDeleteTemplateMethod) == null) {
          SubscriptionServiceGrpc.getDeleteTemplateMethod = getDeleteTemplateMethod =
              io.grpc.MethodDescriptor.<subscription.SubscriptionOuterClass.GetRequest, subscription.SubscriptionOuterClass.Empty>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DeleteTemplate"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.GetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  subscription.SubscriptionOuterClass.Empty.getDefaultInstance()))
              .setSchemaDescriptor(new SubscriptionServiceMethodDescriptorSupplier("DeleteTemplate"))
              .build();
        }
      }
    }
    return getDeleteTemplateMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static SubscriptionServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<SubscriptionServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<SubscriptionServiceStub>() {
        @java.lang.Override
        public SubscriptionServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new SubscriptionServiceStub(channel, callOptions);
        }
      };
    return SubscriptionServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static SubscriptionServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<SubscriptionServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<SubscriptionServiceBlockingStub>() {
        @java.lang.Override
        public SubscriptionServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new SubscriptionServiceBlockingStub(channel, callOptions);
        }
      };
    return SubscriptionServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static SubscriptionServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<SubscriptionServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<SubscriptionServiceFutureStub>() {
        @java.lang.Override
        public SubscriptionServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new SubscriptionServiceFutureStub(channel, callOptions);
        }
      };
    return SubscriptionServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * ============================================================
   * 5. СЕРВИС (главная часть — список всех RPC-методов)
   * ============================================================
   * Service — это интерфейс. Он описывает ВСЕ методы API, которые будут доступны
   * через gRPC.
   * Каждый rpc — это один метод.
   * Формат: rpc &lt;ИмяМетода&gt; (Запрос) returns (Ответ);
   * Имена методов должны быть глаголами (Create, Get, List, Update, Delete).
   * Это общепринятое соглашение для gRPC.
   * !!! ВАЖНО: методы делятся на 2 группы по смыслу:
   *   1. Работа с ПОДПИСКАМИ (CRUD) — для пользователей
   *   2. Работа с ШАБЛОНАМИ (CRUD) — для администраторов
   * Разделение такое же, как в REST-эндпоинтах:
   *   GET/POST /api/subscriptions      → работа с подписками
   *   GET/POST /api/admin/templates    → работа с шаблонами (только админ)
   * </pre>
   */
  public interface AsyncService {

    /**
     * <pre>
     * --- GET /api/subscriptions ---
     * Получить список всех подписок (с пагинацией).
     * На вход: limit (сколько записей), offset (сдвиг от начала).
     * На выход: список подписок (SubscriptionList).
     * В REST это выглядит так:
     *   GET /api/subscriptions?limit=10&amp;offset=0
     * </pre>
     */
    default void getSubscriptions(subscription.SubscriptionOuterClass.GetSubscriptionsRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.SubscriptionList> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetSubscriptionsMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- POST /api/subscriptions ---
     * Создать новую подписку.
     * На вход: template_id, user_id, start_date, end_date.
     * На выход: ID созданной подписки.
     * В REST это выглядит так:
     *   POST /api/subscriptions
     *   Body: { "template_id": 1, "user_id": "...", "start_date": "08-2025", "end_date": "..." }
     * </pre>
     */
    default void createSubscription(subscription.SubscriptionOuterClass.CreateRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.CreateResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateSubscriptionMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions/{id} ---
     * Получить одну подписку по её ID.
     * На вход: id.
     * На выход: полная запись подписки (Subscription).
     * В REST это выглядит так:
     *   GET /api/subscriptions/5
     * </pre>
     */
    default void getSubscription(subscription.SubscriptionOuterClass.GetRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Subscription> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetSubscriptionMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- PUT /api/subscriptions/{id} ---
     * Обновить существующую подписку.
     * На вход: id + новые поля.
     * На выход: ничего (Empty), только статус успеха.
     * В REST это выглядит так:
     *   PUT /api/subscriptions/5
     *   Body: { "template_id": 2, "user_id": "...", "start_date": "09-2025", ... }
     * </pre>
     */
    default void updateSubscription(subscription.SubscriptionOuterClass.UpdateRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateSubscriptionMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- DELETE /api/subscriptions/{id} ---
     * Удалить подписку по ID.
     * На вход: id.
     * На выход: ничего (Empty).
     * В REST это выглядит так:
     *   DELETE /api/subscriptions/5
     * </pre>
     */
    default void deleteSubscription(subscription.SubscriptionOuterClass.GetRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDeleteSubscriptionMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions/total-cost ---
     * Рассчитать суммарную стоимость подписок за период.
     * На вход: user_id, service_name (опционально), start_date, end_date.
     * На выход: total (число).
     * В REST это выглядит так:
     *   GET /api/subscriptions/total-cost?user_id=...&amp;start_date=01-2025&amp;end_date=12-2025
     * </pre>
     */
    default void getTotalCost(subscription.SubscriptionOuterClass.TotalCostRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.TotalCostResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetTotalCostMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- GET /api/templates ---
     * Получить список всех шаблонов (доступно всем авторизованным пользователям).
     * На вход: ничего (Empty).
     * На выход: список шаблонов (TemplateList).
     * </pre>
     */
    default void listTemplates(subscription.SubscriptionOuterClass.Empty request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.TemplateList> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListTemplatesMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- GET /api/templates/{id} ---
     * Получить шаблон по ID.
     * На вход: id.
     * На выход: полная запись шаблона (Template).
     * </pre>
     */
    default void getTemplate(subscription.SubscriptionOuterClass.GetRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Template> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetTemplateMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- POST /api/admin/templates ---
     * Создать новый шаблон (только админ).
     * На вход: service_name, price.
     * На выход: ID созданного шаблона.
     * </pre>
     */
    default void createTemplate(subscription.SubscriptionOuterClass.CreateTemplateRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.CreateResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateTemplateMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- PUT /api/admin/templates/{id} ---
     * Обновить шаблон (только админ).
     * На вход: id + service_name, price.
     * На выход: ничего (Empty).
     * </pre>
     */
    default void updateTemplate(subscription.SubscriptionOuterClass.UpdateTemplateRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateTemplateMethod(), responseObserver);
    }

    /**
     * <pre>
     * --- DELETE /api/admin/templates/{id} ---
     * Удалить шаблон (только админ).
     * На вход: id.
     * На выход: ничего (Empty).
     * </pre>
     */
    default void deleteTemplate(subscription.SubscriptionOuterClass.GetRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDeleteTemplateMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service SubscriptionService.
   * <pre>
   * ============================================================
   * 5. СЕРВИС (главная часть — список всех RPC-методов)
   * ============================================================
   * Service — это интерфейс. Он описывает ВСЕ методы API, которые будут доступны
   * через gRPC.
   * Каждый rpc — это один метод.
   * Формат: rpc &lt;ИмяМетода&gt; (Запрос) returns (Ответ);
   * Имена методов должны быть глаголами (Create, Get, List, Update, Delete).
   * Это общепринятое соглашение для gRPC.
   * !!! ВАЖНО: методы делятся на 2 группы по смыслу:
   *   1. Работа с ПОДПИСКАМИ (CRUD) — для пользователей
   *   2. Работа с ШАБЛОНАМИ (CRUD) — для администраторов
   * Разделение такое же, как в REST-эндпоинтах:
   *   GET/POST /api/subscriptions      → работа с подписками
   *   GET/POST /api/admin/templates    → работа с шаблонами (только админ)
   * </pre>
   */
  public static abstract class SubscriptionServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return SubscriptionServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service SubscriptionService.
   * <pre>
   * ============================================================
   * 5. СЕРВИС (главная часть — список всех RPC-методов)
   * ============================================================
   * Service — это интерфейс. Он описывает ВСЕ методы API, которые будут доступны
   * через gRPC.
   * Каждый rpc — это один метод.
   * Формат: rpc &lt;ИмяМетода&gt; (Запрос) returns (Ответ);
   * Имена методов должны быть глаголами (Create, Get, List, Update, Delete).
   * Это общепринятое соглашение для gRPC.
   * !!! ВАЖНО: методы делятся на 2 группы по смыслу:
   *   1. Работа с ПОДПИСКАМИ (CRUD) — для пользователей
   *   2. Работа с ШАБЛОНАМИ (CRUD) — для администраторов
   * Разделение такое же, как в REST-эндпоинтах:
   *   GET/POST /api/subscriptions      → работа с подписками
   *   GET/POST /api/admin/templates    → работа с шаблонами (только админ)
   * </pre>
   */
  public static final class SubscriptionServiceStub
      extends io.grpc.stub.AbstractAsyncStub<SubscriptionServiceStub> {
    private SubscriptionServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected SubscriptionServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new SubscriptionServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions ---
     * Получить список всех подписок (с пагинацией).
     * На вход: limit (сколько записей), offset (сдвиг от начала).
     * На выход: список подписок (SubscriptionList).
     * В REST это выглядит так:
     *   GET /api/subscriptions?limit=10&amp;offset=0
     * </pre>
     */
    public void getSubscriptions(subscription.SubscriptionOuterClass.GetSubscriptionsRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.SubscriptionList> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetSubscriptionsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- POST /api/subscriptions ---
     * Создать новую подписку.
     * На вход: template_id, user_id, start_date, end_date.
     * На выход: ID созданной подписки.
     * В REST это выглядит так:
     *   POST /api/subscriptions
     *   Body: { "template_id": 1, "user_id": "...", "start_date": "08-2025", "end_date": "..." }
     * </pre>
     */
    public void createSubscription(subscription.SubscriptionOuterClass.CreateRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.CreateResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateSubscriptionMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions/{id} ---
     * Получить одну подписку по её ID.
     * На вход: id.
     * На выход: полная запись подписки (Subscription).
     * В REST это выглядит так:
     *   GET /api/subscriptions/5
     * </pre>
     */
    public void getSubscription(subscription.SubscriptionOuterClass.GetRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Subscription> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetSubscriptionMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- PUT /api/subscriptions/{id} ---
     * Обновить существующую подписку.
     * На вход: id + новые поля.
     * На выход: ничего (Empty), только статус успеха.
     * В REST это выглядит так:
     *   PUT /api/subscriptions/5
     *   Body: { "template_id": 2, "user_id": "...", "start_date": "09-2025", ... }
     * </pre>
     */
    public void updateSubscription(subscription.SubscriptionOuterClass.UpdateRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateSubscriptionMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- DELETE /api/subscriptions/{id} ---
     * Удалить подписку по ID.
     * На вход: id.
     * На выход: ничего (Empty).
     * В REST это выглядит так:
     *   DELETE /api/subscriptions/5
     * </pre>
     */
    public void deleteSubscription(subscription.SubscriptionOuterClass.GetRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDeleteSubscriptionMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions/total-cost ---
     * Рассчитать суммарную стоимость подписок за период.
     * На вход: user_id, service_name (опционально), start_date, end_date.
     * На выход: total (число).
     * В REST это выглядит так:
     *   GET /api/subscriptions/total-cost?user_id=...&amp;start_date=01-2025&amp;end_date=12-2025
     * </pre>
     */
    public void getTotalCost(subscription.SubscriptionOuterClass.TotalCostRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.TotalCostResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetTotalCostMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- GET /api/templates ---
     * Получить список всех шаблонов (доступно всем авторизованным пользователям).
     * На вход: ничего (Empty).
     * На выход: список шаблонов (TemplateList).
     * </pre>
     */
    public void listTemplates(subscription.SubscriptionOuterClass.Empty request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.TemplateList> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListTemplatesMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- GET /api/templates/{id} ---
     * Получить шаблон по ID.
     * На вход: id.
     * На выход: полная запись шаблона (Template).
     * </pre>
     */
    public void getTemplate(subscription.SubscriptionOuterClass.GetRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Template> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetTemplateMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- POST /api/admin/templates ---
     * Создать новый шаблон (только админ).
     * На вход: service_name, price.
     * На выход: ID созданного шаблона.
     * </pre>
     */
    public void createTemplate(subscription.SubscriptionOuterClass.CreateTemplateRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.CreateResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateTemplateMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- PUT /api/admin/templates/{id} ---
     * Обновить шаблон (только админ).
     * На вход: id + service_name, price.
     * На выход: ничего (Empty).
     * </pre>
     */
    public void updateTemplate(subscription.SubscriptionOuterClass.UpdateTemplateRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateTemplateMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * --- DELETE /api/admin/templates/{id} ---
     * Удалить шаблон (только админ).
     * На вход: id.
     * На выход: ничего (Empty).
     * </pre>
     */
    public void deleteTemplate(subscription.SubscriptionOuterClass.GetRequest request,
        io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDeleteTemplateMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service SubscriptionService.
   * <pre>
   * ============================================================
   * 5. СЕРВИС (главная часть — список всех RPC-методов)
   * ============================================================
   * Service — это интерфейс. Он описывает ВСЕ методы API, которые будут доступны
   * через gRPC.
   * Каждый rpc — это один метод.
   * Формат: rpc &lt;ИмяМетода&gt; (Запрос) returns (Ответ);
   * Имена методов должны быть глаголами (Create, Get, List, Update, Delete).
   * Это общепринятое соглашение для gRPC.
   * !!! ВАЖНО: методы делятся на 2 группы по смыслу:
   *   1. Работа с ПОДПИСКАМИ (CRUD) — для пользователей
   *   2. Работа с ШАБЛОНАМИ (CRUD) — для администраторов
   * Разделение такое же, как в REST-эндпоинтах:
   *   GET/POST /api/subscriptions      → работа с подписками
   *   GET/POST /api/admin/templates    → работа с шаблонами (только админ)
   * </pre>
   */
  public static final class SubscriptionServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<SubscriptionServiceBlockingStub> {
    private SubscriptionServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected SubscriptionServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new SubscriptionServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions ---
     * Получить список всех подписок (с пагинацией).
     * На вход: limit (сколько записей), offset (сдвиг от начала).
     * На выход: список подписок (SubscriptionList).
     * В REST это выглядит так:
     *   GET /api/subscriptions?limit=10&amp;offset=0
     * </pre>
     */
    public subscription.SubscriptionOuterClass.SubscriptionList getSubscriptions(subscription.SubscriptionOuterClass.GetSubscriptionsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetSubscriptionsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- POST /api/subscriptions ---
     * Создать новую подписку.
     * На вход: template_id, user_id, start_date, end_date.
     * На выход: ID созданной подписки.
     * В REST это выглядит так:
     *   POST /api/subscriptions
     *   Body: { "template_id": 1, "user_id": "...", "start_date": "08-2025", "end_date": "..." }
     * </pre>
     */
    public subscription.SubscriptionOuterClass.CreateResponse createSubscription(subscription.SubscriptionOuterClass.CreateRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateSubscriptionMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions/{id} ---
     * Получить одну подписку по её ID.
     * На вход: id.
     * На выход: полная запись подписки (Subscription).
     * В REST это выглядит так:
     *   GET /api/subscriptions/5
     * </pre>
     */
    public subscription.SubscriptionOuterClass.Subscription getSubscription(subscription.SubscriptionOuterClass.GetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetSubscriptionMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- PUT /api/subscriptions/{id} ---
     * Обновить существующую подписку.
     * На вход: id + новые поля.
     * На выход: ничего (Empty), только статус успеха.
     * В REST это выглядит так:
     *   PUT /api/subscriptions/5
     *   Body: { "template_id": 2, "user_id": "...", "start_date": "09-2025", ... }
     * </pre>
     */
    public subscription.SubscriptionOuterClass.Empty updateSubscription(subscription.SubscriptionOuterClass.UpdateRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateSubscriptionMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- DELETE /api/subscriptions/{id} ---
     * Удалить подписку по ID.
     * На вход: id.
     * На выход: ничего (Empty).
     * В REST это выглядит так:
     *   DELETE /api/subscriptions/5
     * </pre>
     */
    public subscription.SubscriptionOuterClass.Empty deleteSubscription(subscription.SubscriptionOuterClass.GetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDeleteSubscriptionMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions/total-cost ---
     * Рассчитать суммарную стоимость подписок за период.
     * На вход: user_id, service_name (опционально), start_date, end_date.
     * На выход: total (число).
     * В REST это выглядит так:
     *   GET /api/subscriptions/total-cost?user_id=...&amp;start_date=01-2025&amp;end_date=12-2025
     * </pre>
     */
    public subscription.SubscriptionOuterClass.TotalCostResponse getTotalCost(subscription.SubscriptionOuterClass.TotalCostRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetTotalCostMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- GET /api/templates ---
     * Получить список всех шаблонов (доступно всем авторизованным пользователям).
     * На вход: ничего (Empty).
     * На выход: список шаблонов (TemplateList).
     * </pre>
     */
    public subscription.SubscriptionOuterClass.TemplateList listTemplates(subscription.SubscriptionOuterClass.Empty request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListTemplatesMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- GET /api/templates/{id} ---
     * Получить шаблон по ID.
     * На вход: id.
     * На выход: полная запись шаблона (Template).
     * </pre>
     */
    public subscription.SubscriptionOuterClass.Template getTemplate(subscription.SubscriptionOuterClass.GetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetTemplateMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- POST /api/admin/templates ---
     * Создать новый шаблон (только админ).
     * На вход: service_name, price.
     * На выход: ID созданного шаблона.
     * </pre>
     */
    public subscription.SubscriptionOuterClass.CreateResponse createTemplate(subscription.SubscriptionOuterClass.CreateTemplateRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateTemplateMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- PUT /api/admin/templates/{id} ---
     * Обновить шаблон (только админ).
     * На вход: id + service_name, price.
     * На выход: ничего (Empty).
     * </pre>
     */
    public subscription.SubscriptionOuterClass.Empty updateTemplate(subscription.SubscriptionOuterClass.UpdateTemplateRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateTemplateMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * --- DELETE /api/admin/templates/{id} ---
     * Удалить шаблон (только админ).
     * На вход: id.
     * На выход: ничего (Empty).
     * </pre>
     */
    public subscription.SubscriptionOuterClass.Empty deleteTemplate(subscription.SubscriptionOuterClass.GetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDeleteTemplateMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service SubscriptionService.
   * <pre>
   * ============================================================
   * 5. СЕРВИС (главная часть — список всех RPC-методов)
   * ============================================================
   * Service — это интерфейс. Он описывает ВСЕ методы API, которые будут доступны
   * через gRPC.
   * Каждый rpc — это один метод.
   * Формат: rpc &lt;ИмяМетода&gt; (Запрос) returns (Ответ);
   * Имена методов должны быть глаголами (Create, Get, List, Update, Delete).
   * Это общепринятое соглашение для gRPC.
   * !!! ВАЖНО: методы делятся на 2 группы по смыслу:
   *   1. Работа с ПОДПИСКАМИ (CRUD) — для пользователей
   *   2. Работа с ШАБЛОНАМИ (CRUD) — для администраторов
   * Разделение такое же, как в REST-эндпоинтах:
   *   GET/POST /api/subscriptions      → работа с подписками
   *   GET/POST /api/admin/templates    → работа с шаблонами (только админ)
   * </pre>
   */
  public static final class SubscriptionServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<SubscriptionServiceFutureStub> {
    private SubscriptionServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected SubscriptionServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new SubscriptionServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions ---
     * Получить список всех подписок (с пагинацией).
     * На вход: limit (сколько записей), offset (сдвиг от начала).
     * На выход: список подписок (SubscriptionList).
     * В REST это выглядит так:
     *   GET /api/subscriptions?limit=10&amp;offset=0
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.SubscriptionList> getSubscriptions(
        subscription.SubscriptionOuterClass.GetSubscriptionsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetSubscriptionsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- POST /api/subscriptions ---
     * Создать новую подписку.
     * На вход: template_id, user_id, start_date, end_date.
     * На выход: ID созданной подписки.
     * В REST это выглядит так:
     *   POST /api/subscriptions
     *   Body: { "template_id": 1, "user_id": "...", "start_date": "08-2025", "end_date": "..." }
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.CreateResponse> createSubscription(
        subscription.SubscriptionOuterClass.CreateRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateSubscriptionMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions/{id} ---
     * Получить одну подписку по её ID.
     * На вход: id.
     * На выход: полная запись подписки (Subscription).
     * В REST это выглядит так:
     *   GET /api/subscriptions/5
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.Subscription> getSubscription(
        subscription.SubscriptionOuterClass.GetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetSubscriptionMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- PUT /api/subscriptions/{id} ---
     * Обновить существующую подписку.
     * На вход: id + новые поля.
     * На выход: ничего (Empty), только статус успеха.
     * В REST это выглядит так:
     *   PUT /api/subscriptions/5
     *   Body: { "template_id": 2, "user_id": "...", "start_date": "09-2025", ... }
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.Empty> updateSubscription(
        subscription.SubscriptionOuterClass.UpdateRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateSubscriptionMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- DELETE /api/subscriptions/{id} ---
     * Удалить подписку по ID.
     * На вход: id.
     * На выход: ничего (Empty).
     * В REST это выглядит так:
     *   DELETE /api/subscriptions/5
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.Empty> deleteSubscription(
        subscription.SubscriptionOuterClass.GetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDeleteSubscriptionMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- GET /api/subscriptions/total-cost ---
     * Рассчитать суммарную стоимость подписок за период.
     * На вход: user_id, service_name (опционально), start_date, end_date.
     * На выход: total (число).
     * В REST это выглядит так:
     *   GET /api/subscriptions/total-cost?user_id=...&amp;start_date=01-2025&amp;end_date=12-2025
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.TotalCostResponse> getTotalCost(
        subscription.SubscriptionOuterClass.TotalCostRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetTotalCostMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- GET /api/templates ---
     * Получить список всех шаблонов (доступно всем авторизованным пользователям).
     * На вход: ничего (Empty).
     * На выход: список шаблонов (TemplateList).
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.TemplateList> listTemplates(
        subscription.SubscriptionOuterClass.Empty request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListTemplatesMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- GET /api/templates/{id} ---
     * Получить шаблон по ID.
     * На вход: id.
     * На выход: полная запись шаблона (Template).
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.Template> getTemplate(
        subscription.SubscriptionOuterClass.GetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetTemplateMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- POST /api/admin/templates ---
     * Создать новый шаблон (только админ).
     * На вход: service_name, price.
     * На выход: ID созданного шаблона.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.CreateResponse> createTemplate(
        subscription.SubscriptionOuterClass.CreateTemplateRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateTemplateMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- PUT /api/admin/templates/{id} ---
     * Обновить шаблон (только админ).
     * На вход: id + service_name, price.
     * На выход: ничего (Empty).
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.Empty> updateTemplate(
        subscription.SubscriptionOuterClass.UpdateTemplateRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateTemplateMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * --- DELETE /api/admin/templates/{id} ---
     * Удалить шаблон (только админ).
     * На вход: id.
     * На выход: ничего (Empty).
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<subscription.SubscriptionOuterClass.Empty> deleteTemplate(
        subscription.SubscriptionOuterClass.GetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDeleteTemplateMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_GET_SUBSCRIPTIONS = 0;
  private static final int METHODID_CREATE_SUBSCRIPTION = 1;
  private static final int METHODID_GET_SUBSCRIPTION = 2;
  private static final int METHODID_UPDATE_SUBSCRIPTION = 3;
  private static final int METHODID_DELETE_SUBSCRIPTION = 4;
  private static final int METHODID_GET_TOTAL_COST = 5;
  private static final int METHODID_LIST_TEMPLATES = 6;
  private static final int METHODID_GET_TEMPLATE = 7;
  private static final int METHODID_CREATE_TEMPLATE = 8;
  private static final int METHODID_UPDATE_TEMPLATE = 9;
  private static final int METHODID_DELETE_TEMPLATE = 10;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_GET_SUBSCRIPTIONS:
          serviceImpl.getSubscriptions((subscription.SubscriptionOuterClass.GetSubscriptionsRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.SubscriptionList>) responseObserver);
          break;
        case METHODID_CREATE_SUBSCRIPTION:
          serviceImpl.createSubscription((subscription.SubscriptionOuterClass.CreateRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.CreateResponse>) responseObserver);
          break;
        case METHODID_GET_SUBSCRIPTION:
          serviceImpl.getSubscription((subscription.SubscriptionOuterClass.GetRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Subscription>) responseObserver);
          break;
        case METHODID_UPDATE_SUBSCRIPTION:
          serviceImpl.updateSubscription((subscription.SubscriptionOuterClass.UpdateRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty>) responseObserver);
          break;
        case METHODID_DELETE_SUBSCRIPTION:
          serviceImpl.deleteSubscription((subscription.SubscriptionOuterClass.GetRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty>) responseObserver);
          break;
        case METHODID_GET_TOTAL_COST:
          serviceImpl.getTotalCost((subscription.SubscriptionOuterClass.TotalCostRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.TotalCostResponse>) responseObserver);
          break;
        case METHODID_LIST_TEMPLATES:
          serviceImpl.listTemplates((subscription.SubscriptionOuterClass.Empty) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.TemplateList>) responseObserver);
          break;
        case METHODID_GET_TEMPLATE:
          serviceImpl.getTemplate((subscription.SubscriptionOuterClass.GetRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Template>) responseObserver);
          break;
        case METHODID_CREATE_TEMPLATE:
          serviceImpl.createTemplate((subscription.SubscriptionOuterClass.CreateTemplateRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.CreateResponse>) responseObserver);
          break;
        case METHODID_UPDATE_TEMPLATE:
          serviceImpl.updateTemplate((subscription.SubscriptionOuterClass.UpdateTemplateRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty>) responseObserver);
          break;
        case METHODID_DELETE_TEMPLATE:
          serviceImpl.deleteTemplate((subscription.SubscriptionOuterClass.GetRequest) request,
              (io.grpc.stub.StreamObserver<subscription.SubscriptionOuterClass.Empty>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getGetSubscriptionsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.GetSubscriptionsRequest,
              subscription.SubscriptionOuterClass.SubscriptionList>(
                service, METHODID_GET_SUBSCRIPTIONS)))
        .addMethod(
          getCreateSubscriptionMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.CreateRequest,
              subscription.SubscriptionOuterClass.CreateResponse>(
                service, METHODID_CREATE_SUBSCRIPTION)))
        .addMethod(
          getGetSubscriptionMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.GetRequest,
              subscription.SubscriptionOuterClass.Subscription>(
                service, METHODID_GET_SUBSCRIPTION)))
        .addMethod(
          getUpdateSubscriptionMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.UpdateRequest,
              subscription.SubscriptionOuterClass.Empty>(
                service, METHODID_UPDATE_SUBSCRIPTION)))
        .addMethod(
          getDeleteSubscriptionMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.GetRequest,
              subscription.SubscriptionOuterClass.Empty>(
                service, METHODID_DELETE_SUBSCRIPTION)))
        .addMethod(
          getGetTotalCostMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.TotalCostRequest,
              subscription.SubscriptionOuterClass.TotalCostResponse>(
                service, METHODID_GET_TOTAL_COST)))
        .addMethod(
          getListTemplatesMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.Empty,
              subscription.SubscriptionOuterClass.TemplateList>(
                service, METHODID_LIST_TEMPLATES)))
        .addMethod(
          getGetTemplateMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.GetRequest,
              subscription.SubscriptionOuterClass.Template>(
                service, METHODID_GET_TEMPLATE)))
        .addMethod(
          getCreateTemplateMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.CreateTemplateRequest,
              subscription.SubscriptionOuterClass.CreateResponse>(
                service, METHODID_CREATE_TEMPLATE)))
        .addMethod(
          getUpdateTemplateMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.UpdateTemplateRequest,
              subscription.SubscriptionOuterClass.Empty>(
                service, METHODID_UPDATE_TEMPLATE)))
        .addMethod(
          getDeleteTemplateMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              subscription.SubscriptionOuterClass.GetRequest,
              subscription.SubscriptionOuterClass.Empty>(
                service, METHODID_DELETE_TEMPLATE)))
        .build();
  }

  private static abstract class SubscriptionServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    SubscriptionServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return subscription.SubscriptionOuterClass.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("SubscriptionService");
    }
  }

  private static final class SubscriptionServiceFileDescriptorSupplier
      extends SubscriptionServiceBaseDescriptorSupplier {
    SubscriptionServiceFileDescriptorSupplier() {}
  }

  private static final class SubscriptionServiceMethodDescriptorSupplier
      extends SubscriptionServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    SubscriptionServiceMethodDescriptorSupplier(java.lang.String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (SubscriptionServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new SubscriptionServiceFileDescriptorSupplier())
              .addMethod(getGetSubscriptionsMethod())
              .addMethod(getCreateSubscriptionMethod())
              .addMethod(getGetSubscriptionMethod())
              .addMethod(getUpdateSubscriptionMethod())
              .addMethod(getDeleteSubscriptionMethod())
              .addMethod(getGetTotalCostMethod())
              .addMethod(getListTemplatesMethod())
              .addMethod(getGetTemplateMethod())
              .addMethod(getCreateTemplateMethod())
              .addMethod(getUpdateTemplateMethod())
              .addMethod(getDeleteTemplateMethod())
              .build();
        }
      }
    }
    return result;
  }
}
